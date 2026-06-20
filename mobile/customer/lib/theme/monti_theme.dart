import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';

import 'tokens.dart';

/// Builds the [ThemeData] used app-wide. Inter font per DES-0008 §1.2.
class MontiTheme {
  const MontiTheme._();

  static ThemeData dark() {
    final base = ThemeData.dark(useMaterial3: true);
    final textTheme = GoogleFonts.interTextTheme(base.textTheme).apply(
      bodyColor: MontiColors.textPrimary,
      displayColor: MontiColors.textPrimary,
    );

    return base.copyWith(
      scaffoldBackgroundColor: MontiColors.bgBase,
      colorScheme: const ColorScheme.dark(
        surface: MontiColors.bgSurface,
        primary: MontiColors.accentCyan,
        onPrimary: MontiColors.bgBase,
        secondary: MontiColors.accentCyanSoft,
        error: MontiColors.danger,
      ),
      textTheme: textTheme,
      appBarTheme: const AppBarTheme(
        backgroundColor: MontiColors.bgBase,
        elevation: 0,
        centerTitle: false,
        iconTheme: IconThemeData(color: MontiColors.textPrimary),
        titleTextStyle: TextStyle(
          color: MontiColors.textPrimary,
          fontSize: 18,
          fontWeight: FontWeight.w600,
        ),
      ),
      cardTheme: CardThemeData(
        color: MontiColors.bgSurface,
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(MontiRadii.card),
        ),
      ),
      elevatedButtonTheme: ElevatedButtonThemeData(
        style: ElevatedButton.styleFrom(
          backgroundColor: MontiColors.accentCyan,
          foregroundColor: MontiColors.bgBase,
          minimumSize: const Size.fromHeight(MontiSpacing.tapTargetMin),
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(MontiRadii.card),
          ),
          textStyle: const TextStyle(
            fontSize: 16,
            fontWeight: FontWeight.w600,
            letterSpacing: 0.4,
          ),
        ),
      ),
      outlinedButtonTheme: OutlinedButtonThemeData(
        style: OutlinedButton.styleFrom(
          foregroundColor: MontiColors.accentCyan,
          side: const BorderSide(color: MontiColors.accentCyan, width: 1.5),
          minimumSize: const Size.fromHeight(MontiSpacing.tapTargetMin),
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(MontiRadii.card),
          ),
          textStyle: const TextStyle(
            fontSize: 16,
            fontWeight: FontWeight.w600,
            letterSpacing: 0.4,
          ),
        ),
      ),
      focusColor: MontiColors.accentCyan,
    );
  }
}
