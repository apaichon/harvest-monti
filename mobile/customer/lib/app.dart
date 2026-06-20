import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import 'l10n/app_localizations.dart';
import 'models/menu.dart';
import 'screens/01_welcome.dart';
import 'screens/02_language.dart';
import 'screens/03_choose_category.dart';
import 'screens/04_menu_list.dart';
import 'screens/05_item_detail.dart';
import 'screens/06_cart.dart';
import 'screens/07_checkout_payment.dart';
import 'screens/08_thank_you.dart';
import 'screens/09_order_status.dart';
import 'screens/qr_scan.dart';
import 'state/providers.dart';
import 'theme/monti_theme.dart';

/// Root Monti customer app.
class MontiApp extends ConsumerWidget {
  const MontiApp({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final locale = ref.watch(localeProvider);
    return MaterialApp.router(
      title: 'Monti',
      debugShowCheckedModeBanner: false,
      theme: MontiTheme.dark(),
      locale: locale,
      supportedLocales: AppLocalizations.supportedLocales,
      localizationsDelegates: AppLocalizations.localizationsDelegates,
      routerConfig: _router,
    );
  }
}

final _router = GoRouter(
  initialLocation: '/',
  routes: [
    GoRoute(path: '/', builder: (_, __) => const WelcomeScreen()),
    GoRoute(path: '/qr', builder: (_, __) => const QrScanScreen()),
    GoRoute(path: '/language', builder: (_, __) => const LanguageSelectionScreen()),
    GoRoute(path: '/categories', builder: (_, __) => const ChooseCategoryScreen()),
    GoRoute(
      path: '/menu/:catId',
      builder: (_, s) => MenuListScreen(categoryId: s.pathParameters['catId']!),
    ),
    GoRoute(
      path: '/item/:id',
      builder: (_, s) {
        final extra = s.extra;
        if (extra is MenuItem) return ItemDetailScreen(item: extra);
        // Fallback stub if user deep-linked.
        return ItemDetailScreen(
          item: MenuItem(id: s.pathParameters['id']!, name: '—', priceTHB: 0),
        );
      },
    ),
    GoRoute(path: '/cart', builder: (_, __) => const CartScreen()),
    GoRoute(path: '/checkout', builder: (_, __) => const CheckoutPaymentScreen()),
    GoRoute(
      path: '/thank-you/:orderNumber',
      builder: (_, s) =>
          ThankYouScreen(orderNumber: s.pathParameters['orderNumber']!),
    ),
    GoRoute(
      path: '/order/:orderNumber',
      builder: (_, s) =>
          OrderStatusScreen(orderNumber: s.pathParameters['orderNumber']!),
    ),
  ],
);
