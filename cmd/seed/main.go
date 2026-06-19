// Command seed bootstraps the MONTI demo tenant + menu + placeholder WebP
// images on a single run. Re-running the binary is a no-op — every INSERT
// uses ON CONFLICT DO NOTHING and every MinIO PutObject is gated by
// StatObject. See TASK-0013 / DES-0007 §1, §4.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/apaichon/harvest-monti/cmd/seed/data"
	"github.com/apaichon/harvest-monti/cmd/seed/images"
)

const (
	// MinIO bucket layout per DES-0007 §4.
	menuImagesBucket = "monti-menu-images"
	// 1y immutable cache header per DES-0007 §4.
	cacheControlImmutable = "public, max-age=31536000, immutable"
	imageContentType      = "image/webp"
)

// requiredTables — if any of these is missing we abort before touching
// Postgres or MinIO so the operator gets a single, actionable message.
var requiredTables = []string{
	"tenants", "outlets", "menus", "menu_categories", "menu_items",
	"modifier_groups", "modifier_options", "allergens", "item_allergens",
	"promo_codes",
}

func main() {
	if err := run(context.Background()); err != nil {
		log.Fatalf("seed: %v", err)
	}
}

func run(ctx context.Context) error {
	dryRun := os.Getenv("TEST_NO_INFRA") == "1"

	dsn := os.Getenv("DATABASE_URL")
	if !dryRun && dsn == "" {
		return errors.New("DATABASE_URL is required (set TEST_NO_INFRA=1 for dry-run)")
	}
	mEndpoint := os.Getenv("MINIO_ENDPOINT")
	mAccessKey := os.Getenv("MINIO_ACCESS_KEY")
	mSecretKey := os.Getenv("MINIO_SECRET_KEY")
	if !dryRun && (mEndpoint == "" || mAccessKey == "" || mSecretKey == "") {
		return errors.New("MINIO_ENDPOINT/MINIO_ACCESS_KEY/MINIO_SECRET_KEY are required (set TEST_NO_INFRA=1 for dry-run)")
	}

	seed := data.Build()
	logSeedSummary(seed)
	log.Printf("seed: webp encoder = %s", images.Encoder)

	if dryRun {
		log.Print("seed: TEST_NO_INFRA=1 — dry-run only; skipping DB writes and MinIO uploads")
		return dryRunRenderProbe(seed)
	}

	// Postgres
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return fmt.Errorf("postgres connect: %w", err)
	}
	defer conn.Close(ctx)

	if err := assertSchema(ctx, conn); err != nil {
		return err
	}

	// MinIO
	mc, err := newMinioClient(mEndpoint, mAccessKey, mSecretKey)
	if err != nil {
		return fmt.Errorf("minio client: %w", err)
	}
	if err := assertBucket(ctx, mc, menuImagesBucket); err != nil {
		return err
	}

	if err := writeAllergens(ctx, conn, seed.Allergens); err != nil {
		return fmt.Errorf("allergens: %w", err)
	}
	if err := writeTenant(ctx, conn, seed.Tenant); err != nil {
		return fmt.Errorf("tenant: %w", err)
	}
	if err := writeOutlet(ctx, conn, seed.Outlet); err != nil {
		return fmt.Errorf("outlet: %w", err)
	}
	if err := writeMenu(ctx, conn, seed.Menu); err != nil {
		return fmt.Errorf("menu: %w", err)
	}
	for _, c := range seed.Categories {
		if err := writeCategory(ctx, conn, c, seed.Tenant.ID); err != nil {
			return fmt.Errorf("category %s: %w", c.Name, err)
		}
	}
	for _, it := range seed.Items {
		if err := writeItem(ctx, conn, it, seed.Tenant.ID); err != nil {
			return fmt.Errorf("item %s: %w", it.Name, err)
		}
	}
	for _, mg := range seed.Modifiers {
		if err := writeModifierGroup(ctx, conn, mg); err != nil {
			return fmt.Errorf("modifier group %s: %w", mg.Name, err)
		}
	}
	if err := writeItemAllergens(ctx, conn, seed); err != nil {
		return fmt.Errorf("item_allergens: %w", err)
	}
	for _, p := range seed.Promos {
		if err := writePromo(ctx, conn, p); err != nil {
			return fmt.Errorf("promo %s: %w", p.Code, err)
		}
	}

	// MinIO uploads
	if err := uploadCategoryImages(ctx, mc, seed); err != nil {
		return fmt.Errorf("category images: %w", err)
	}
	if err := uploadItemImages(ctx, mc, seed); err != nil {
		return fmt.Errorf("item images: %w", err)
	}

	log.Print("seed: done")
	return nil
}

func logSeedSummary(s data.Seed) {
	log.Printf("seed: tenant=%s (%s) outlet=%s menu=%q categories=%d items=%d modifier_groups=%d allergens=%d promos=%d",
		s.Tenant.Name, s.Tenant.ID, s.Outlet.Name, s.Menu.Name,
		len(s.Categories), len(s.Items), len(s.Modifiers), len(s.Allergens), len(s.Promos))
}

// dryRunRenderProbe encodes each image once locally so we can record the
// resulting byte count in the report without touching infra. It does NOT
// reach Postgres or MinIO.
func dryRunRenderProbe(s data.Seed) error {
	for _, c := range s.Categories {
		b, err := images.Generate(480, 480, c.Name, c.AccentColor)
		if err != nil {
			return fmt.Errorf("category %s: %w", c.Name, err)
		}
		log.Printf("seed: [dry-run] category %-7s webp=%d bytes", c.Name, len(b))
	}
	for _, it := range s.Items {
		b, err := images.Generate(1200, 675, it.Name, it.AccentColor)
		if err != nil {
			return fmt.Errorf("item %s: %w", it.Name, err)
		}
		log.Printf("seed: [dry-run] item %-25s webp=%d bytes", it.Name, len(b))
	}
	return nil
}

// ---------------- Postgres helpers ----------------

func assertSchema(ctx context.Context, conn *pgx.Conn) error {
	for _, t := range requiredTables {
		var exists bool
		err := conn.QueryRow(ctx, `SELECT EXISTS (
			SELECT 1 FROM information_schema.tables
			WHERE table_schema = 'public' AND table_name = $1)`, t).Scan(&exists)
		if err != nil {
			return fmt.Errorf("schema probe %s: %w", t, err)
		}
		if !exists {
			return fmt.Errorf("required table %q is missing — run migrations first (TASK-0009 / cmd/migrate up)", t)
		}
	}
	return nil
}

func writeTenant(ctx context.Context, conn *pgx.Conn, t data.Tenant) error {
	// tenants has no natural unique key beyond id; ON CONFLICT (id) DO NOTHING.
	_, err := conn.Exec(ctx, `
		INSERT INTO tenants (id, name, locale)
		VALUES ($1, $2, $3)
		ON CONFLICT (id) DO NOTHING`, t.ID, t.Name, t.Locale)
	return err
}

func writeOutlet(ctx context.Context, conn *pgx.Conn, o data.Outlet) error {
	_, err := conn.Exec(ctx, `
		INSERT INTO outlets (id, tenant_id, name, table_count, timezone)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO NOTHING`,
		o.ID, o.TenantID, o.Name, o.TableCount, o.Timezone)
	return err
}

func writeMenu(ctx context.Context, conn *pgx.Conn, m data.Menu) error {
	t, err := time.Parse("2006-01-02", m.EffectiveFrom)
	if err != nil {
		return fmt.Errorf("parse effective_from: %w", err)
	}
	_, err = conn.Exec(ctx, `
		INSERT INTO menus (id, outlet_id, name, status, effective_from)
		VALUES ($1, $2, $3, $4::menu_status, $5)
		ON CONFLICT (id) DO NOTHING`,
		m.ID, m.OutletID, m.Name, m.Status, t)
	return err
}

func writeCategory(ctx context.Context, conn *pgx.Conn, c data.Category, tenantID uuid.UUID) error {
	imgPath := fmt.Sprintf("%s/%s/categories/%s.webp", menuImagesBucket, tenantID, c.ID)
	_, err := conn.Exec(ctx, `
		INSERT INTO menu_categories (id, menu_id, name, display_order, image_minio_path, status)
		VALUES ($1, $2, $3, $4, $5, 'active'::category_status)
		ON CONFLICT (id) DO NOTHING`,
		c.ID, c.MenuID, c.Name, c.DisplayOrder, imgPath)
	return err
}

func writeItem(ctx context.Context, conn *pgx.Conn, it data.Item, tenantID uuid.UUID) error {
	imgPath := fmt.Sprintf("%s/%s/%s.webp", menuImagesBucket, tenantID, it.ID)
	_, err := conn.Exec(ctx, `
		INSERT INTO menu_items (
			id, menu_id, category_id, sku, name, description,
			price_cents, currency, image_minio_path, is_best_seller,
			display_order, status
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, 'active'::item_status)
		ON CONFLICT (id) DO NOTHING`,
		it.ID, it.MenuID, it.CategoryID, it.SKU, it.Name, it.Description,
		it.PriceCents, it.Currency, imgPath, it.IsBestSeller, it.DisplayOrder)
	return err
}

func writeModifierGroup(ctx context.Context, conn *pgx.Conn, mg data.ModifierGroup) error {
	tx, err := conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	_, err = tx.Exec(ctx, `
		INSERT INTO modifier_groups (id, item_id, name, kind, required, min_select, max_select)
		VALUES ($1, $2, $3, $4::modifier_kind, $5, $6, $7)
		ON CONFLICT (id) DO NOTHING`,
		mg.ID, mg.ItemID, mg.Name, mg.Kind, mg.Required, mg.MinSelect, mg.MaxSelect)
	if err != nil {
		return err
	}
	for _, o := range mg.Options {
		_, err = tx.Exec(ctx, `
			INSERT INTO modifier_options (id, group_id, name, price_delta_cents, display_order)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (id) DO NOTHING`,
			o.ID, mg.ID, o.Name, o.PriceDeltaCents, o.DisplayOrder)
		if err != nil {
			return fmt.Errorf("option %s: %w", o.Name, err)
		}
	}
	return tx.Commit(ctx)
}

func writeAllergens(ctx context.Context, conn *pgx.Conn, allergens []data.Allergen) error {
	for _, a := range allergens {
		// allergens.code has a unique index — ON CONFLICT on code keeps the
		// original row (and original id) on re-run.
		_, err := conn.Exec(ctx, `
			INSERT INTO allergens (id, code, label_en, label_th, label_zh, label_ja)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (code) DO NOTHING`,
			a.ID, a.Code, a.LabelEN, a.LabelTH, a.LabelZH, a.LabelJA)
		if err != nil {
			return fmt.Errorf("allergen %s: %w", a.Code, err)
		}
	}
	return nil
}

func writeItemAllergens(ctx context.Context, conn *pgx.Conn, s data.Seed) error {
	// Resolve allergen codes -> ids from the DB so we link by code regardless
	// of which row owns the canonical id (defends against a previous seed run
	// that inserted with a different id).
	rows, err := conn.Query(ctx, `SELECT code, id FROM allergens`)
	if err != nil {
		return err
	}
	codeToID := map[string]uuid.UUID{}
	for rows.Next() {
		var code string
		var id uuid.UUID
		if err := rows.Scan(&code, &id); err != nil {
			rows.Close()
			return err
		}
		codeToID[code] = id
	}
	rows.Close()

	for itemID, codes := range s.ItemAllergens {
		for _, code := range codes {
			aid, ok := codeToID[code]
			if !ok {
				return fmt.Errorf("allergen code %q not found in allergens table", code)
			}
			_, err := conn.Exec(ctx, `
				INSERT INTO item_allergens (item_id, allergen_id)
				VALUES ($1, $2)
				ON CONFLICT (item_id, allergen_id) DO NOTHING`, itemID, aid)
			if err != nil {
				return fmt.Errorf("item %s + allergen %s: %w", itemID, code, err)
			}
		}
	}
	return nil
}

func writePromo(ctx context.Context, conn *pgx.Conn, p data.PromoCode) error {
	var expires any
	if p.ExpiresAt != "" {
		t, err := time.Parse("2006-01-02", p.ExpiresAt)
		if err != nil {
			return fmt.Errorf("parse expires_at: %w", err)
		}
		expires = t
	}
	_, err := conn.Exec(ctx, `
		INSERT INTO promo_codes (id, tenant_id, code, kind, value, max_uses, expires_at, status)
		VALUES ($1, $2, $3, $4::promo_kind, $5, $6, $7, 'active'::promo_status)
		ON CONFLICT (tenant_id, code) DO NOTHING`,
		p.ID, p.TenantID, p.Code, p.Kind, p.Value, p.MaxUses, expires)
	if pgErr := asPGError(err); pgErr != nil && pgErr.Code == "23505" {
		// belt-and-braces — per-tenant unique index conflict handled above.
		return nil
	}
	return err
}

func asPGError(err error) *pgconn.PgError {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr
	}
	return nil
}

// ---------------- MinIO helpers ----------------

func newMinioClient(endpoint, ak, sk string) (*minio.Client, error) {
	// Allow http://host:port form; minio-go expects host:port + Secure flag.
	secure := true
	host := endpoint
	if u, err := url.Parse(endpoint); err == nil && (u.Scheme == "http" || u.Scheme == "https") {
		host = u.Host
		secure = u.Scheme == "https"
	}
	return minio.New(host, &minio.Options{
		Creds:  credentials.NewStaticV4(ak, sk, ""),
		Secure: secure,
	})
}

func assertBucket(ctx context.Context, mc *minio.Client, bucket string) error {
	ok, err := mc.BucketExists(ctx, bucket)
	if err != nil {
		return fmt.Errorf("BucketExists %q: %w", bucket, err)
	}
	if !ok {
		return fmt.Errorf("bucket %q is missing — run TASK-0009 MinIO bootstrap first", bucket)
	}
	return nil
}

func uploadIfMissing(ctx context.Context, mc *minio.Client, bucket, key string, payload []byte) error {
	_, err := mc.StatObject(ctx, bucket, key, minio.StatObjectOptions{})
	if err == nil {
		log.Printf("seed: minio object already present, skipping: %s/%s", bucket, key)
		return nil
	}
	if !isNotFound(err) {
		return fmt.Errorf("StatObject %q: %w", key, err)
	}
	_, err = mc.PutObject(ctx, bucket, key,
		bytesReader(payload), int64(len(payload)),
		minio.PutObjectOptions{
			ContentType:  imageContentType,
			CacheControl: cacheControlImmutable,
		})
	if err != nil {
		return fmt.Errorf("PutObject %q: %w", key, err)
	}
	log.Printf("seed: uploaded %s/%s (%d bytes)", bucket, key, len(payload))
	return nil
}

func isNotFound(err error) bool {
	if err == nil {
		return false
	}
	resp := minio.ToErrorResponse(err)
	if resp.Code == "NoSuchKey" || resp.StatusCode == 404 {
		return true
	}
	// Fallback string match for non-S3 implementations.
	return strings.Contains(strings.ToLower(err.Error()), "not exist") ||
		strings.Contains(strings.ToLower(err.Error()), "not found")
}

func bytesReader(b []byte) *strings.Reader {
	// strings.Reader satisfies io.Reader without an extra allocation path.
	return strings.NewReader(string(b))
}

func uploadCategoryImages(ctx context.Context, mc *minio.Client, s data.Seed) error {
	for _, c := range s.Categories {
		img, err := images.Generate(480, 480, c.Name, c.AccentColor)
		if err != nil {
			return fmt.Errorf("render category %s: %w", c.Name, err)
		}
		key := fmt.Sprintf("%s/categories/%s.webp", s.Tenant.ID, c.ID)
		if err := uploadIfMissing(ctx, mc, menuImagesBucket, key, img); err != nil {
			return err
		}
	}
	return nil
}

func uploadItemImages(ctx context.Context, mc *minio.Client, s data.Seed) error {
	for _, it := range s.Items {
		img, err := images.Generate(1200, 675, it.Name, it.AccentColor)
		if err != nil {
			return fmt.Errorf("render item %s: %w", it.Name, err)
		}
		key := fmt.Sprintf("%s/%s.webp", s.Tenant.ID, it.ID)
		if err := uploadIfMissing(ctx, mc, menuImagesBucket, key, img); err != nil {
			return err
		}
	}
	return nil
}
