import 'package:flutter/material.dart';

/// Design tokens — mirrors DES-0008 §1.1 palette and §1.2 typography.
///
/// Keep this file purely declarative. Anything that needs a [BuildContext]
/// belongs in [MontiTheme].
class MontiColors {
  const MontiColors._();

  static const Color bgBase = Color(0xFF0B1438);
  static const Color bgSurface = Color(0xFF142051);
  static const Color bgSurface2 = Color(0xFF1B2A66);
  static const Color accentCyan = Color(0xFF00BFFF);
  static const Color accentCyanSoft = Color(0xAA33CCFF);
  static const Color textPrimary = Color(0xFFFFFFFF);
  static const Color textSecondary = Color(0x99FFFFFF); // 60%
  static const Color textDisabled = Color(0x61FFFFFF); // 38%
  static const Color success = Color(0xFF22C55E);
  static const Color warn = Color(0xFFF59E0B);
  static const Color danger = Color(0xFFEF4444);
}

class MontiRadii {
  const MontiRadii._();

  /// rounded-2xl per DES-0008 §1.1
  static const double card = 24.0;

  /// rounded-3xl per DES-0008 §1.1
  static const double rail = 32.0;

  /// Small chips / buttons
  static const double chip = 12.0;
}

class MontiSpacing {
  const MontiSpacing._();

  static const double xs = 4.0;
  static const double sm = 8.0;
  static const double md = 16.0;
  static const double lg = 24.0;
  static const double xl = 32.0;

  /// Mobile minimum tap target per DES-0008 §1.6
  static const double tapTargetMin = 48.0;
}
