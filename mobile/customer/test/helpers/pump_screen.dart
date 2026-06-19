import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'package:monti_customer/l10n/app_localizations.dart';
import 'package:monti_customer/theme/monti_theme.dart';

/// Wraps a widget in ProviderScope + MaterialApp with localizations so
/// each screen-under-test can be pumped in isolation.
Widget pumpScreen(
  Widget child, {
  List<Override> overrides = const [],
  Locale locale = const Locale('en'),
}) {
  return ProviderScope(
    overrides: overrides,
    child: MaterialApp(
      locale: locale,
      theme: MontiTheme.dark(),
      supportedLocales: AppLocalizations.supportedLocales,
      localizationsDelegates: AppLocalizations.localizationsDelegates,
      home: child,
    ),
  );
}
