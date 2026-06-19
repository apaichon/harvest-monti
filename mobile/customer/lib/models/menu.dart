import 'package:flutter/foundation.dart';

/// Menu category (DES-0008 §3.3 tile).
@immutable
class MenuCategory {
  const MenuCategory({
    required this.id,
    required this.name,
    this.imageUrl,
  });

  final String id;
  final String name;
  final String? imageUrl;

  factory MenuCategory.fromJson(Map<String, dynamic> json) => MenuCategory(
        id: json['id'] as String,
        name: json['name'] as String,
        imageUrl: json['image_url'] as String?,
      );
}

/// A modifier choice within a [ModifierGroup].
@immutable
class ModifierOption {
  const ModifierOption({
    required this.id,
    required this.label,
    this.priceDelta = 0,
  });

  final String id;
  final String label;
  final int priceDelta;

  factory ModifierOption.fromJson(Map<String, dynamic> json) => ModifierOption(
        id: json['id'] as String,
        label: json['label'] as String,
        priceDelta: (json['price_delta'] ?? 0) as int,
      );
}

/// Modifier group rendering rules per DES-0008 §7.3.
enum ModifierSelectionKind { single, multi }

@immutable
class ModifierGroup {
  const ModifierGroup({
    required this.id,
    required this.label,
    required this.selectionKind,
    required this.required,
    required this.options,
  });

  final String id;
  final String label;
  final ModifierSelectionKind selectionKind;
  final bool required;
  final List<ModifierOption> options;

  factory ModifierGroup.fromJson(Map<String, dynamic> json) => ModifierGroup(
        id: json['id'] as String,
        label: json['label'] as String,
        selectionKind: (json['selection_kind'] as String) == 'multi'
            ? ModifierSelectionKind.multi
            : ModifierSelectionKind.single,
        required: (json['required'] ?? false) as bool,
        options: ((json['options'] ?? const []) as List)
            .map((o) => ModifierOption.fromJson(o as Map<String, dynamic>))
            .toList(growable: false),
      );
}

/// Menu item — see DES-0008 §3.4/§3.5.
@immutable
class MenuItem {
  const MenuItem({
    required this.id,
    required this.name,
    required this.priceTHB,
    this.description = '',
    this.imageUrl,
    this.badges = const [],
    this.categoryId,
    this.modifierGroups = const [],
  });

  final String id;
  final String name;
  final String description;
  final int priceTHB;
  final String? imageUrl;
  final List<String> badges;
  final String? categoryId;
  final List<ModifierGroup> modifierGroups;

  factory MenuItem.fromJson(Map<String, dynamic> json) => MenuItem(
        id: json['id'] as String,
        name: json['name'] as String,
        description: (json['description'] ?? '') as String,
        priceTHB: (json['price_thb'] ?? json['price'] ?? 0) as int,
        imageUrl: json['image_url'] as String?,
        badges: ((json['badges'] ?? const []) as List).cast<String>(),
        categoryId: json['category_id'] as String?,
        modifierGroups: ((json['modifier_groups'] ?? const []) as List)
            .map((m) => ModifierGroup.fromJson(m as Map<String, dynamic>))
            .toList(growable: false),
      );
}
