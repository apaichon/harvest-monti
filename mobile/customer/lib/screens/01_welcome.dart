import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../l10n/app_localizations.dart';
import '../theme/tokens.dart';
import '../widgets/monti_mascot.dart';
import '../widgets/voice_mic_fab.dart';

/// DES-0008 §3.1 — Welcome screen.
class WelcomeScreen extends ConsumerWidget {
  const WelcomeScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l = AppLocalizations.of(context);
    return Scaffold(
      appBar: AppBar(automaticallyImplyLeading: false),
      floatingActionButton: const VoiceMicFab(),
      body: SafeArea(
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: MontiSpacing.lg),
          child: Column(
            children: [
              const SizedBox(height: MontiSpacing.xl),
              const MontiMascot(),
              const SizedBox(height: MontiSpacing.lg),
              Text(
                l.welcomeGreeting,
                style: const TextStyle(
                  fontSize: 28,
                  fontWeight: FontWeight.w700,
                  color: MontiColors.textPrimary,
                ),
              ),
              const SizedBox(height: MontiSpacing.sm),
              Text(
                l.welcomePrompt,
                style: const TextStyle(
                  fontSize: 16,
                  color: MontiColors.textSecondary,
                ),
                textAlign: TextAlign.center,
              ),
              const Spacer(),
              SizedBox(
                width: double.infinity,
                child: ElevatedButton(
                  key: const Key('welcome.start_order'),
                  onPressed: () => context.go('/language'),
                  child: Text(l.startOrder),
                ),
              ),
              const SizedBox(height: MontiSpacing.md),
              SizedBox(
                width: double.infinity,
                child: OutlinedButton.icon(
                  key: const Key('welcome.scan_qr'),
                  onPressed: () => context.go('/qr'),
                  icon: const Icon(Icons.camera_alt_outlined),
                  label: Text(l.scanQrCode),
                ),
              ),
              const SizedBox(height: MontiSpacing.lg),
              Text(
                l.footerVersion,
                style: const TextStyle(
                  color: MontiColors.textSecondary,
                  fontSize: 12,
                ),
              ),
              const SizedBox(height: MontiSpacing.md),
            ],
          ),
        ),
      ),
    );
  }
}
