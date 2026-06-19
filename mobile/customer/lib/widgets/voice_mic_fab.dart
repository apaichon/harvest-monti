import 'package:flutter/material.dart';

import '../theme/tokens.dart';

/// Floating mic FAB present on every screen per REQ-0011 / task spec.
///
/// Tapping opens a voice session against `/api/v1/voice/sessions` — wired in
/// the Voice intro screen for the demo; this widget is the persistent
/// affordance per REQ-0010 AC-12.
class VoiceMicFab extends StatelessWidget {
  const VoiceMicFab({super.key, this.onPressed});

  final VoidCallback? onPressed;

  @override
  Widget build(BuildContext context) {
    return FloatingActionButton(
      heroTag: 'voice_mic_fab',
      backgroundColor: MontiColors.accentCyan,
      foregroundColor: MontiColors.bgBase,
      onPressed: onPressed ?? () {},
      tooltip: 'Talk to Monti',
      child: const Icon(Icons.mic),
    );
  }
}
