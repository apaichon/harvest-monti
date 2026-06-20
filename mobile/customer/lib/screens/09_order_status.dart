import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../l10n/app_localizations.dart';
import '../state/order_status_controller.dart';
import '../state/providers.dart';
import '../theme/tokens.dart';
import '../widgets/order_status_stepper.dart';
import '../widgets/voice_mic_fab.dart';

/// DES-0008 §3.9 — Order Status with WS subscribe + 10s poll fallback.
class OrderStatusScreen extends ConsumerStatefulWidget {
  const OrderStatusScreen({super.key, required this.orderNumber});
  final String orderNumber;

  @override
  ConsumerState<OrderStatusScreen> createState() => _OrderStatusScreenState();
}

class _OrderStatusScreenState extends ConsumerState<OrderStatusScreen> {
  @override
  void initState() {
    super.initState();
    // Kick off the WS subscription on first frame.
    WidgetsBinding.instance.addPostFrameCallback((_) {
      ref.read(orderStatusProvider(widget.orderNumber).notifier).connect();
    });
  }

  @override
  Widget build(BuildContext context) {
    final l = AppLocalizations.of(context);
    final state = ref.watch(orderStatusProvider(widget.orderNumber));
    return Scaffold(
      appBar: AppBar(title: Text(l.orderStatusTitle)),
      floatingActionButton: const VoiceMicFab(),
      body: SafeArea(
        child: Padding(
          padding: const EdgeInsets.all(MontiSpacing.lg),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                '${l.orderNumberLabel} ${state.status.orderNumber}',
                style: const TextStyle(
                    fontSize: 18, fontWeight: FontWeight.w700),
              ),
              const SizedBox(height: MontiSpacing.md),
              Container(
                padding: const EdgeInsets.symmetric(
                    horizontal: MontiSpacing.md, vertical: MontiSpacing.sm),
                decoration: BoxDecoration(
                  color: MontiColors.bgSurface,
                  borderRadius: BorderRadius.circular(MontiRadii.card),
                ),
                child: Text(
                  '${l.etaLabel}  ${state.status.etaMinLow}-${state.status.etaMinHigh} min',
                  style: const TextStyle(
                      color: MontiColors.accentCyan,
                      fontWeight: FontWeight.w600),
                ),
              ),
              const SizedBox(height: MontiSpacing.xl),
              OrderStatusStepper(
                currentStep: state.status.currentStep,
                receivedAtLocal: state.status.receivedAtLocal ?? '12:04 PM',
                labels: (
                  received: l.stepReceived,
                  preparing: l.stepPreparing,
                  ready: l.stepReady,
                ),
              ),
              const Spacer(),
              if (state.connection == ConnectionMode.polling)
                const Padding(
                  padding: EdgeInsets.only(bottom: MontiSpacing.sm),
                  child: Text(
                    'live updates paused — refreshing every 10s',
                    style: TextStyle(
                        color: MontiColors.textSecondary, fontSize: 12),
                  ),
                ),
              SizedBox(
                width: double.infinity,
                child: OutlinedButton.icon(
                  key: const Key('status.call_staff'),
                  onPressed: () {},
                  icon: const Icon(Icons.room_service_outlined),
                  label: Text(l.callStaff),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
