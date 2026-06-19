import 'package:flutter/material.dart';

import '../models/order.dart';
import '../theme/tokens.dart';

/// DES-0008 §3.9 — vertical 3-step progression with leading dot/check/ring.
class OrderStatusStepper extends StatelessWidget {
  const OrderStatusStepper({
    super.key,
    required this.currentStep,
    required this.labels,
    this.receivedAtLocal,
  });

  final OrderStep currentStep;

  /// (received, preparing, ready) localized labels.
  final ({String received, String preparing, String ready}) labels;
  final String? receivedAtLocal;

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        _row(
          icon: const Icon(Icons.check, color: MontiColors.success, size: 20),
          label: labels.received,
          sub: receivedAtLocal,
          active: true,
        ),
        const _Connector(),
        _row(
          icon: currentStep == OrderStep.preparing
              ? const SizedBox(
                  width: 20,
                  height: 20,
                  child: CircularProgressIndicator(
                    color: MontiColors.accentCyan,
                    strokeWidth: 2.5,
                  ),
                )
              : Icon(
                  Icons.circle,
                  size: 12,
                  color: currentStep.index >= OrderStep.preparing.index
                      ? MontiColors.accentCyan
                      : MontiColors.textDisabled,
                ),
          label: labels.preparing,
          sub: currentStep == OrderStep.preparing ? 'in progress' : null,
          active: currentStep.index >= OrderStep.preparing.index,
        ),
        const _Connector(),
        _row(
          icon: Icon(
            currentStep == OrderStep.ready
                ? Icons.check_circle
                : Icons.radio_button_unchecked,
            color: currentStep == OrderStep.ready
                ? MontiColors.success
                : MontiColors.textDisabled,
            size: 20,
          ),
          label: labels.ready,
          sub: currentStep == OrderStep.ready ? null : 'pending',
          active: currentStep == OrderStep.ready,
        ),
      ],
    );
  }

  Widget _row({
    required Widget icon,
    required String label,
    String? sub,
    required bool active,
  }) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: MontiSpacing.sm),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.center,
        children: [
          SizedBox(width: 28, child: Center(child: icon)),
          const SizedBox(width: MontiSpacing.md),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  label,
                  style: TextStyle(
                    color: active
                        ? MontiColors.textPrimary
                        : MontiColors.textSecondary,
                    fontSize: 16,
                    fontWeight: FontWeight.w600,
                  ),
                ),
                if (sub != null)
                  Text(
                    sub,
                    style: const TextStyle(
                      color: MontiColors.textSecondary,
                      fontSize: 13,
                    ),
                  ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

class _Connector extends StatelessWidget {
  const _Connector();
  @override
  Widget build(BuildContext context) => Padding(
        padding: const EdgeInsets.only(left: 13),
        child: Container(
          width: 2,
          height: 20,
          color: MontiColors.bgSurface2,
        ),
      );
}
