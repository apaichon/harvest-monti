import 'package:flutter/material.dart';

import '../models/order.dart';
import '../theme/tokens.dart';

/// DES-0008 §3.7 payment tile — radio card.
class PaymentMethodTile extends StatelessWidget {
  const PaymentMethodTile({
    super.key,
    required this.method,
    required this.label,
    required this.icon,
    required this.selected,
    required this.onSelect,
  });

  final PaymentMethod method;
  final String label;
  final IconData icon;
  final bool selected;
  final ValueChanged<PaymentMethod> onSelect;

  @override
  Widget build(BuildContext context) {
    return Semantics(
      button: true,
      selected: selected,
      label: label,
      child: InkWell(
        borderRadius: BorderRadius.circular(MontiRadii.card),
        onTap: () => onSelect(method),
        child: Container(
          padding: const EdgeInsets.all(MontiSpacing.md),
          decoration: BoxDecoration(
            color: selected ? MontiColors.bgSurface2 : MontiColors.bgSurface,
            borderRadius: BorderRadius.circular(MontiRadii.card),
            border: Border.all(
              color: selected ? MontiColors.accentCyan : Colors.transparent,
              width: 2,
            ),
          ),
          child: Row(
            children: [
              Icon(icon, color: MontiColors.accentCyan, size: 28),
              const SizedBox(width: MontiSpacing.md),
              Expanded(
                child: Text(
                  label,
                  style: const TextStyle(
                    color: MontiColors.textPrimary,
                    fontSize: 16,
                    fontWeight: FontWeight.w600,
                  ),
                ),
              ),
              Radio<bool>(
                value: true,
                groupValue: selected ? true : null,
                onChanged: (_) => onSelect(method),
                activeColor: MontiColors.accentCyan,
              ),
            ],
          ),
        ),
      ),
    );
  }
}
