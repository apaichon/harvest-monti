import 'package:flutter/material.dart';

import '../theme/tokens.dart';

/// Placeholder for the Monti mascot illustration.
///
/// DES-0008 §1.4 calls for an SVG mascot with a 64px radial cyan glow at 30%
/// opacity. Until the SVG/Rive asset lands, we render a deterministic
/// geometric stand-in that preserves layout footprint and golden tolerance.
class MontiMascot extends StatelessWidget {
  const MontiMascot({super.key, this.size = 160, this.glow = true});

  final double size;
  final bool glow;

  @override
  Widget build(BuildContext context) {
    return Semantics(
      label: 'Monti mascot',
      child: Container(
        width: size,
        height: size,
        decoration: BoxDecoration(
          shape: BoxShape.circle,
          color: MontiColors.bgSurface,
          boxShadow: glow
              ? const [
                  BoxShadow(
                    color: MontiColors.accentCyanSoft,
                    blurRadius: 64,
                    spreadRadius: 4,
                  ),
                ]
              : null,
        ),
        child: const Center(
          child: Icon(
            Icons.pets,
            color: MontiColors.accentCyan,
            size: 72,
          ),
        ),
      ),
    );
  }
}
