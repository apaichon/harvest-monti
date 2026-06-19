import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:mobile_scanner/mobile_scanner.dart';

import '../l10n/app_localizations.dart';
import '../state/providers.dart';
import '../theme/tokens.dart';
import '../widgets/voice_mic_fab.dart';

/// Camera-backed QR scan with manual-entry fallback per TASK-0012 edge case
/// "camera permission denied".
class QrScanScreen extends ConsumerStatefulWidget {
  const QrScanScreen({super.key});
  @override
  ConsumerState<QrScanScreen> createState() => _QrScanScreenState();
}

class _QrScanScreenState extends ConsumerState<QrScanScreen> {
  bool _busy = false;
  String? _error;
  final _manualCtrl = TextEditingController();

  @override
  void dispose() {
    _manualCtrl.dispose();
    super.dispose();
  }

  Future<void> _redeem(String token) async {
    if (_busy) return;
    setState(() => _busy = true);
    try {
      final api = ref.read(qrApiProvider);
      final result = await api.redeem(token);
      ref.read(tableCodeProvider.notifier).state = result.tableCode;
      ref.read(sessionIdProvider.notifier).state = result.sessionId;
      if (!mounted) return;
      context.go('/categories');
    } catch (e) {
      setState(() {
        _error = '$e';
        _busy = false;
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    final l = AppLocalizations.of(context);
    return Scaffold(
      appBar: AppBar(title: const Text('QR')),
      floatingActionButton: const VoiceMicFab(),
      body: SafeArea(
        child: Padding(
          padding: const EdgeInsets.all(MontiSpacing.lg),
          child: Column(
            children: [
              Expanded(
                child: ClipRRect(
                  borderRadius: BorderRadius.circular(MontiRadii.card),
                  child: MobileScanner(
                    onDetect: (capture) {
                      final raw = capture.barcodes.firstOrNull?.rawValue;
                      if (raw != null) _redeem(raw);
                    },
                    errorBuilder: (context, _) => _ManualEntry(
                      controller: _manualCtrl,
                      onSubmit: _redeem,
                      hint: l.tableCodeHint,
                      submitLabel: l.submit,
                      errorLabel: l.cameraPermissionDenied,
                    ),
                  ),
                ),
              ),
              if (_error != null)
                Padding(
                  padding: const EdgeInsets.only(top: MontiSpacing.sm),
                  child: Text(_error!,
                      style: const TextStyle(color: MontiColors.danger)),
                ),
              const SizedBox(height: MontiSpacing.md),
              SizedBox(
                width: double.infinity,
                child: OutlinedButton(
                  key: const Key('qr.manual_entry'),
                  onPressed: () => showModalBottomSheet<void>(
                    context: context,
                    backgroundColor: MontiColors.bgSurface,
                    builder: (_) => _ManualEntry(
                      controller: _manualCtrl,
                      onSubmit: _redeem,
                      hint: l.tableCodeHint,
                      submitLabel: l.submit,
                      errorLabel: '',
                    ),
                  ),
                  child: Text(l.manualTableEntry),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _ManualEntry extends StatelessWidget {
  const _ManualEntry({
    required this.controller,
    required this.onSubmit,
    required this.hint,
    required this.submitLabel,
    required this.errorLabel,
  });
  final TextEditingController controller;
  final ValueChanged<String> onSubmit;
  final String hint;
  final String submitLabel;
  final String errorLabel;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.all(MontiSpacing.lg),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          if (errorLabel.isNotEmpty)
            Text(errorLabel,
                style: const TextStyle(color: MontiColors.danger)),
          const SizedBox(height: MontiSpacing.md),
          TextField(
            key: const Key('qr.manual_input'),
            controller: controller,
            decoration: InputDecoration(hintText: hint),
          ),
          const SizedBox(height: MontiSpacing.md),
          SizedBox(
            width: double.infinity,
            child: ElevatedButton(
              key: const Key('qr.manual_submit'),
              onPressed: () => onSubmit(controller.text.trim()),
              child: Text(submitLabel),
            ),
          ),
        ],
      ),
    );
  }
}

// Note: Iterable<T>.firstOrNull ships with Dart 3 collection extensions.
