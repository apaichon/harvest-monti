import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../l10n/app_localizations.dart';
import '../models/cart.dart';
import '../state/providers.dart';
import '../theme/tokens.dart';
import '../widgets/voice_mic_fab.dart';

/// DES-0008 §3.6 — Cart.
class CartScreen extends ConsumerStatefulWidget {
  const CartScreen({super.key});
  @override
  ConsumerState<CartScreen> createState() => _CartScreenState();
}

class _CartScreenState extends ConsumerState<CartScreen> {
  final _promoCtrl = TextEditingController();

  @override
  void dispose() {
    _promoCtrl.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final l = AppLocalizations.of(context);
    final cart = ref.watch(cartProvider);
    final totals = cart.totals;

    return Scaffold(
      appBar: AppBar(title: Text(l.yourCart)),
      floatingActionButton: const VoiceMicFab(),
      body: SafeArea(
        child: Padding(
          padding: const EdgeInsets.all(MontiSpacing.lg),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              Expanded(
                child: ListView.separated(
                  itemCount: cart.lines.length,
                  separatorBuilder: (_, __) =>
                      const Divider(height: 1, color: MontiColors.bgSurface2),
                  itemBuilder: (_, i) {
                    final line = cart.lines[i];
                    return _LineRow(
                      line: line,
                      onMinus: () => ref
                          .read(cartProvider.notifier)
                          .updateQty(line.lineId, line.qty - 1),
                      onPlus: () => ref
                          .read(cartProvider.notifier)
                          .updateQty(line.lineId, line.qty + 1),
                      onRemove: () => ref
                          .read(cartProvider.notifier)
                          .removeLine(line.lineId),
                    );
                  },
                ),
              ),
              const SizedBox(height: MontiSpacing.md),
              Row(children: [
                Expanded(
                  child: TextField(
                    key: const Key('cart.promo'),
                    controller: _promoCtrl,
                    decoration: InputDecoration(
                      hintText: l.promoCodeHint,
                      filled: true,
                      fillColor: MontiColors.bgSurface,
                      border: OutlineInputBorder(
                        borderRadius: BorderRadius.circular(MontiRadii.card),
                        borderSide: BorderSide.none,
                      ),
                    ),
                  ),
                ),
                const SizedBox(width: MontiSpacing.sm),
                IconButton(
                  key: const Key('cart.promo_apply'),
                  icon: const Icon(Icons.check_circle,
                      color: MontiColors.accentCyan),
                  onPressed: () {
                    final code = _promoCtrl.text.trim().toUpperCase();
                    if (code.isEmpty) return;
                    // Local stub — server reconciles per REQ-0012 AC-5.
                    ref.read(cartProvider.notifier).applyPromo(
                          AppliedPromo(code: code, discountTHB: 50),
                        );
                  },
                ),
              ]),
              const SizedBox(height: MontiSpacing.md),
              _row(l.subtotal, '฿ ${totals.subtotalTHB}'),
              _row(l.serviceCharge, '฿ ${totals.serviceChargeTHB}'),
              _row(l.tax7Pct, '฿ ${totals.taxTHB}'),
              if (totals.hasDiscount)
                _row(l.discount, '−฿ ${totals.discountTHB}',
                    isDiscount: true),
              const Divider(color: MontiColors.bgSurface2),
              _row(l.total, '฿ ${totals.totalTHB}', emphasis: true),
              const SizedBox(height: MontiSpacing.md),
              SizedBox(
                width: double.infinity,
                child: ElevatedButton(
                  key: const Key('cart.checkout'),
                  onPressed: cart.isEmpty
                      ? null
                      : () => context.go('/checkout'),
                  child: Text(l.checkout),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _row(String label, String value,
      {bool emphasis = false, bool isDiscount = false}) {
    final style = TextStyle(
      fontSize: emphasis ? 18 : 14,
      fontWeight: emphasis ? FontWeight.w700 : FontWeight.w500,
      color: isDiscount ? MontiColors.success : MontiColors.textPrimary,
    );
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [Text(label, style: style), Text(value, style: style)],
      ),
    );
  }
}

class _LineRow extends StatelessWidget {
  const _LineRow({
    required this.line,
    required this.onMinus,
    required this.onPlus,
    required this.onRemove,
  });
  final CartLine line;
  final VoidCallback onMinus;
  final VoidCallback onPlus;
  final VoidCallback onRemove;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: MontiSpacing.sm),
      child: Row(children: [
        Container(
          width: 48,
          height: 48,
          decoration: BoxDecoration(
            color: MontiColors.bgSurface,
            borderRadius: BorderRadius.circular(MontiRadii.chip),
          ),
          child: const Icon(Icons.fastfood, color: MontiColors.accentCyan),
        ),
        const SizedBox(width: MontiSpacing.md),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(line.item.name,
                  style: const TextStyle(fontWeight: FontWeight.w600)),
              const SizedBox(height: 4),
              Row(children: [
                IconButton(
                  key: Key('cart.minus.${line.lineId}'),
                  iconSize: 20,
                  visualDensity: VisualDensity.compact,
                  onPressed: onMinus,
                  icon: const Icon(Icons.remove_circle_outline),
                ),
                Text('${line.qty}'),
                IconButton(
                  key: Key('cart.plus.${line.lineId}'),
                  iconSize: 20,
                  visualDensity: VisualDensity.compact,
                  onPressed: onPlus,
                  icon: const Icon(Icons.add_circle_outline),
                ),
                const SizedBox(width: MontiSpacing.sm),
                Text('฿ ${line.lineTotalTHB}',
                    style: const TextStyle(
                        color: MontiColors.textSecondary, fontSize: 13)),
              ]),
            ],
          ),
        ),
        IconButton(
          key: Key('cart.remove.${line.lineId}'),
          onPressed: onRemove,
          icon: const Icon(Icons.close, color: MontiColors.danger),
        ),
      ]),
    );
  }
}
