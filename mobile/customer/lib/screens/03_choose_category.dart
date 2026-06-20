import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../l10n/app_localizations.dart';
import '../models/menu.dart';
import '../state/providers.dart';
import '../theme/tokens.dart';
import '../widgets/voice_mic_fab.dart';

/// DES-0008 §3.3 — Categories. 2x3 grid loaded from `GET /api/v1/menu`.
class ChooseCategoryScreen extends ConsumerWidget {
  const ChooseCategoryScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l = AppLocalizations.of(context);
    return Scaffold(
      appBar: AppBar(title: Text(l.categoriesTitle)),
      floatingActionButton: const VoiceMicFab(),
      body: SafeArea(
        child: Padding(
          padding: const EdgeInsets.all(MontiSpacing.lg),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                l.chooseCategory,
                style: const TextStyle(
                  fontSize: 22,
                  fontWeight: FontWeight.w700,
                ),
              ),
              const SizedBox(height: MontiSpacing.lg),
              Expanded(
                child: FutureBuilder<List<MenuCategory>>(
                  future: ref.read(menuApiProvider).fetchCategories().catchError(
                    (_) => _fallbackCategories(),
                  ),
                  builder: (context, snap) {
                    final items = snap.data ?? _fallbackCategories();
                    return GridView.builder(
                      gridDelegate:
                          const SliverGridDelegateWithFixedCrossAxisCount(
                        crossAxisCount: 2,
                        mainAxisSpacing: MontiSpacing.md,
                        crossAxisSpacing: MontiSpacing.md,
                        childAspectRatio: 1.0,
                      ),
                      itemCount: items.length,
                      itemBuilder: (_, i) => _CategoryTile(
                        category: items[i],
                        onTap: () => context.go('/menu/${items[i].id}'),
                      ),
                    );
                  },
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

List<MenuCategory> _fallbackCategories() => const [
      MenuCategory(id: 'hot', name: 'Hot Drinks'),
      MenuCategory(id: 'cold', name: 'Cold Drinks'),
      MenuCategory(id: 'pizza', name: 'Pizza'),
      MenuCategory(id: 'pasta', name: 'Pasta'),
      MenuCategory(id: 'burger', name: 'Burger'),
      MenuCategory(id: 'salads', name: 'Salads'),
    ];

class _CategoryTile extends StatelessWidget {
  const _CategoryTile({required this.category, required this.onTap});
  final MenuCategory category;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return InkWell(
      key: Key('category.${category.id}'),
      borderRadius: BorderRadius.circular(MontiRadii.card),
      onTap: onTap,
      child: Container(
        decoration: BoxDecoration(
          color: MontiColors.bgSurface,
          borderRadius: BorderRadius.circular(MontiRadii.card),
        ),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            const Icon(
              Icons.image_outlined,
              size: 48,
              color: MontiColors.accentCyan,
            ),
            const SizedBox(height: MontiSpacing.sm),
            Text(
              category.name,
              style: const TextStyle(
                fontSize: 16,
                fontWeight: FontWeight.w600,
              ),
              textAlign: TextAlign.center,
            ),
          ],
        ),
      ),
    );
  }
}
