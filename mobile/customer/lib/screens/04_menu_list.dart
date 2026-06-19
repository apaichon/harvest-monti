import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../l10n/app_localizations.dart';
import '../models/menu.dart';
import '../state/providers.dart';
import '../theme/tokens.dart';
import '../widgets/voice_mic_fab.dart';

/// DES-0008 §3.4 — Menu List filtered by category with search.
class MenuListScreen extends ConsumerStatefulWidget {
  const MenuListScreen({super.key, required this.categoryId});
  final String categoryId;

  @override
  ConsumerState<MenuListScreen> createState() => _MenuListScreenState();
}

class _MenuListScreenState extends ConsumerState<MenuListScreen> {
  String _query = '';

  @override
  Widget build(BuildContext context) {
    final l = AppLocalizations.of(context);
    final cartCount = ref.watch(cartProvider).lines.length;
    return Scaffold(
      appBar: AppBar(
        title: Text(widget.categoryId),
        actions: [
          IconButton(
            key: const Key('menu.cart_button'),
            icon: Stack(children: [
              const Icon(Icons.shopping_cart_outlined),
              if (cartCount > 0)
                Positioned(
                  right: 0,
                  top: 0,
                  child: Container(
                    padding: const EdgeInsets.all(2),
                    decoration: const BoxDecoration(
                      color: MontiColors.accentCyan,
                      shape: BoxShape.circle,
                    ),
                    constraints: const BoxConstraints(
                      minWidth: 16,
                      minHeight: 16,
                    ),
                    child: Text(
                      '$cartCount',
                      style: const TextStyle(
                        color: MontiColors.bgBase,
                        fontSize: 10,
                        fontWeight: FontWeight.w700,
                      ),
                      textAlign: TextAlign.center,
                    ),
                  ),
                ),
            ]),
            onPressed: () => context.go('/cart'),
          ),
        ],
      ),
      floatingActionButton: const VoiceMicFab(),
      body: SafeArea(
        child: Column(
          children: [
            Padding(
              padding: const EdgeInsets.all(MontiSpacing.md),
              child: TextField(
                key: const Key('menu.search'),
                onChanged: (v) => setState(() => _query = v.toLowerCase()),
                decoration: InputDecoration(
                  prefixIcon: const Icon(Icons.search),
                  hintText: l.menuListSearchHint,
                  filled: true,
                  fillColor: MontiColors.bgSurface,
                  border: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(MontiRadii.card),
                    borderSide: BorderSide.none,
                  ),
                ),
              ),
            ),
            Expanded(
              child: FutureBuilder<List<MenuItem>>(
                future: ref
                    .read(menuApiProvider)
                    .fetchItems(categoryId: widget.categoryId)
                    .catchError((_) => _fallbackItems(widget.categoryId)),
                builder: (context, snap) {
                  final all = snap.data ?? _fallbackItems(widget.categoryId);
                  final items = _query.isEmpty
                      ? all
                      : all
                          .where((i) =>
                              i.name.toLowerCase().contains(_query))
                          .toList();
                  return ListView.separated(
                    itemCount: items.length,
                    separatorBuilder: (_, __) =>
                        const Divider(height: 1, color: MontiColors.bgSurface2),
                    itemBuilder: (_, i) => _MenuRow(
                      item: items[i],
                      onTap: () => context.go('/item/${items[i].id}',
                          extra: items[i]),
                    ),
                  );
                },
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _MenuRow extends StatelessWidget {
  const _MenuRow({required this.item, required this.onTap});
  final MenuItem item;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return InkWell(
      key: Key('menu.row.${item.id}'),
      onTap: onTap,
      child: Padding(
        padding: const EdgeInsets.all(MontiSpacing.md),
        child: Row(
          children: [
            Container(
              width: 56,
              height: 56,
              decoration: BoxDecoration(
                color: MontiColors.bgSurface,
                borderRadius: BorderRadius.circular(MontiRadii.chip),
              ),
              child: const Icon(Icons.fastfood,
                  color: MontiColors.accentCyan),
            ),
            const SizedBox(width: MontiSpacing.md),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    item.name,
                    style: const TextStyle(
                      fontSize: 16,
                      fontWeight: FontWeight.w600,
                    ),
                  ),
                  const SizedBox(height: 4),
                  Row(children: [
                    for (final b in item.badges)
                      Padding(
                        padding: const EdgeInsets.only(right: 6),
                        child: Container(
                          padding: const EdgeInsets.symmetric(
                              horizontal: 6, vertical: 2),
                          decoration: BoxDecoration(
                            color: MontiColors.bgSurface2,
                            borderRadius: BorderRadius.circular(6),
                          ),
                          child: Text(
                            b,
                            style: const TextStyle(
                              fontSize: 10,
                              color: MontiColors.accentCyan,
                              fontWeight: FontWeight.w700,
                            ),
                          ),
                        ),
                      ),
                    Text(
                      '฿ ${item.priceTHB}',
                      style: const TextStyle(
                        color: MontiColors.textSecondary,
                        fontSize: 13,
                      ),
                    ),
                  ]),
                ],
              ),
            ),
            IconButton(
              icon: const Icon(Icons.add_circle,
                  color: MontiColors.accentCyan, size: 32),
              onPressed: onTap,
            ),
          ],
        ),
      ),
    );
  }
}

List<MenuItem> _fallbackItems(String categoryId) => [
      MenuItem(
        id: 'margherita',
        name: 'Margherita Pizza',
        description: 'Tomato, mozzarella, basil.',
        priceTHB: 240,
        badges: const ['BEST'],
        categoryId: categoryId,
        modifierGroups: const [
          ModifierGroup(
            id: 'side',
            label: 'Add Side',
            selectionKind: ModifierSelectionKind.single,
            required: true,
            options: [
              ModifierOption(id: 'salad', label: 'Salad'),
              ModifierOption(id: 'soup', label: 'Soup'),
              ModifierOption(id: 'bread', label: 'Bread'),
            ],
          ),
        ],
      ),
      const MenuItem(
        id: 'pepperoni',
        name: 'Pepperoni Pizza',
        priceTHB: 260,
        categoryId: 'pizza',
      ),
      const MenuItem(
        id: 'hawaiian',
        name: 'Hawaiian Pizza',
        priceTHB: 250,
        categoryId: 'pizza',
      ),
      const MenuItem(
        id: 'quattro',
        name: 'Quattro Formaggi',
        priceTHB: 290,
        badges: ['NEW'],
        categoryId: 'pizza',
      ),
    ];
