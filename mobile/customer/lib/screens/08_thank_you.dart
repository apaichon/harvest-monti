import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../l10n/app_localizations.dart';
import '../theme/tokens.dart';
import '../widgets/voice_mic_fab.dart';

/// DES-0008 §3.8 — Thank You.
class ThankYouScreen extends ConsumerWidget {
  const ThankYouScreen({super.key, required this.orderNumber});
  final String orderNumber;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l = AppLocalizations.of(context);
    return Scaffold(
      floatingActionButton: const VoiceMicFab(),
      body: SafeArea(
        child: Padding(
          padding: const EdgeInsets.all(MontiSpacing.lg),
          child: Column(
            children: [
              const Spacer(),
              Container(
                width: 96,
                height: 96,
                decoration: const BoxDecoration(
                  color: MontiColors.success,
                  shape: BoxShape.circle,
                ),
                child: const Icon(Icons.check,
                    color: Colors.white, size: 56),
              ),
              const SizedBox(height: MontiSpacing.lg),
              Text(l.thankYou,
                  style: const TextStyle(
                      fontSize: 28, fontWeight: FontWeight.w700)),
              const SizedBox(height: MontiSpacing.sm),
              Text(
                l.orderPlacedSuccess,
                style: const TextStyle(
                    color: MontiColors.textSecondary, fontSize: 14),
                textAlign: TextAlign.center,
              ),
              const SizedBox(height: MontiSpacing.lg),
              Text(
                '${l.orderNumberLabel} $orderNumber',
                key: const Key('thank.order_number'),
                style: const TextStyle(
                    fontSize: 16, fontWeight: FontWeight.w600),
              ),
              const Spacer(),
              SizedBox(
                width: double.infinity,
                child: ElevatedButton(
                  key: const Key('thank.continue'),
                  onPressed: () => context.go('/order/$orderNumber'),
                  child: Text(l.continueAction),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
