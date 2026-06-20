import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../l10n/app_localizations.dart';
import '../models/menu.dart';
import '../state/providers.dart';
import '../theme/tokens.dart';
import '../widgets/voice_mic_fab.dart';

/// DES-0008 §3.5 — Item detail with required modifier group + qty stepper.
class ItemDetailScreen extends ConsumerStatefulWidget {
  const ItemDetailScreen({super.key, required this.item});
  final MenuItem item;

  @override
  ConsumerState<ItemDetailScreen> createState() => _ItemDetailScreenState();
}

class _ItemDetailScreenState extends ConsumerState<ItemDetailScreen> {
  int _qty = 1;
  // groupId -> set of optionIds selected.
  final Map<String, Set<String>> _selections = {};

  int get _unitPrice {
    var base = widget.item.priceTHB;
    for (final g in widget.item.modifierGroups) {
      final picked = _selections[g.id] ?? const <String>{};
      for (final o in g.options) {
        if (picked.contains(o.id)) base += o.priceDelta;
      }
    }
    return base;
  }

  bool get _requiredSatisfied {
    for (final g in widget.item.modifierGroups) {
      if (!g.required) continue;
      final picked = _selections[g.id] ?? const <String>{};
      if (picked.isEmpty) return false;
    }
    return true;
  }

  @override
  Widget build(BuildContext context) {
    final l = AppLocalizations.of(context);
    final total = _unitPrice * _qty;
    return Scaffold(
      appBar: AppBar(
        actions: [
          IconButton(
            onPressed: () {},
            icon: const Icon(Icons.favorite_border),
          ),
        ],
      ),
      floatingActionButton: const VoiceMicFab(),
      body: SafeArea(
        child: SingleChildScrollView(
          padding: const EdgeInsets.all(MontiSpacing.lg),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              AspectRatio(
                aspectRatio: 16 / 9,
                child: Container(
                  decoration: BoxDecoration(
                    color: MontiColors.bgSurface,
                    borderRadius: BorderRadius.circular(MontiRadii.card),
                  ),
                  child: const Icon(
                    Icons.image,
                    color: MontiColors.accentCyan,
                    size: 64,
                  ),
                ),
              ),
              const SizedBox(height: MontiSpacing.lg),
              Text(
                widget.item.name,
                style: const TextStyle(
                  fontSize: 24,
                  fontWeight: FontWeight.w700,
                ),
              ),
              const SizedBox(height: MontiSpacing.sm),
              if (widget.item.description.isNotEmpty)
                Text(
                  widget.item.description,
                  style: const TextStyle(
                    color: MontiColors.textSecondary,
                    fontSize: 14,
                  ),
                ),
              const SizedBox(height: MontiSpacing.lg),
              for (final g in widget.item.modifierGroups) _group(g, l),
              const SizedBox(height: MontiSpacing.lg),
              Text(l.quantity,
                  style: const TextStyle(fontWeight: FontWeight.w600)),
              const SizedBox(height: MontiSpacing.sm),
              Row(children: [
                IconButton(
                  key: const Key('item.qty_minus'),
                  onPressed: _qty > 1 ? () => setState(() => _qty--) : null,
                  icon: const Icon(Icons.remove_circle_outline),
                ),
                SizedBox(
                  width: 32,
                  child: Text('$_qty',
                      textAlign: TextAlign.center,
                      style: const TextStyle(
                          fontSize: 18, fontWeight: FontWeight.w700)),
                ),
                IconButton(
                  key: const Key('item.qty_plus'),
                  onPressed: () => setState(() => _qty++),
                  icon: const Icon(Icons.add_circle_outline),
                ),
              ]),
              const SizedBox(height: MontiSpacing.lg),
              SizedBox(
                width: double.infinity,
                child: ElevatedButton(
                  key: const Key('item.add_to_cart'),
                  onPressed: _requiredSatisfied
                      ? () {
                          final flat = <String>{
                            for (final s in _selections.values) ...s,
                          };
                          ref.read(cartProvider.notifier).addLine(
                                item: widget.item,
                                qty: _qty,
                                unitPriceTHB: _unitPrice,
                                selectedModifierOptionIds: flat,
                              );
                          context.go('/cart');
                        }
                      : null,
                  child: Text('${l.addToCart}  ( ฿ $total )'),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _group(ModifierGroup g, AppLocalizations l) {
    return Padding(
      padding: const EdgeInsets.only(bottom: MontiSpacing.md),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            '${g.label}${g.required ? '  (${l.required})' : ''}',
            style: const TextStyle(fontWeight: FontWeight.w600, fontSize: 16),
          ),
          const SizedBox(height: MontiSpacing.sm),
          if (g.selectionKind == ModifierSelectionKind.single)
            ..._singleGroup(g)
          else
            ..._multiGroup(g),
        ],
      ),
    );
  }

  Iterable<Widget> _singleGroup(ModifierGroup g) sync* {
    final picked = _selections[g.id]?.firstOrNull;
    for (final o in g.options) {
      yield RadioListTile<String>(
        key: Key('item.mod.${g.id}.${o.id}'),
        value: o.id,
        groupValue: picked,
        onChanged: (v) => setState(() {
          if (v != null) _selections[g.id] = {v};
        }),
        title: Text(o.label),
        activeColor: MontiColors.accentCyan,
        dense: true,
        contentPadding: EdgeInsets.zero,
      );
    }
  }

  Iterable<Widget> _multiGroup(ModifierGroup g) sync* {
    final picked = _selections.putIfAbsent(g.id, () => <String>{});
    for (final o in g.options) {
      yield CheckboxListTile(
        key: Key('item.mod.${g.id}.${o.id}'),
        value: picked.contains(o.id),
        onChanged: (v) => setState(() {
          if (v == true) {
            picked.add(o.id);
          } else {
            picked.remove(o.id);
          }
        }),
        title: Text(o.label),
        activeColor: MontiColors.accentCyan,
        dense: true,
        contentPadding: EdgeInsets.zero,
      );
    }
  }
}

// Note: Iterable<T>.firstOrNull ships with Dart 3 collection extensions.
