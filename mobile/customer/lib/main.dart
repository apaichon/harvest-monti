import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'app.dart';
import 'state/providers.dart';

Future<void> main() async {
  WidgetsFlutterBinding.ensureInitialized();
  final container = ProviderContainer();
  await container.read(localeProvider.notifier).hydrate();
  runApp(
    UncontrolledProviderScope(
      container: container,
      child: const MontiApp(),
    ),
  );
}
