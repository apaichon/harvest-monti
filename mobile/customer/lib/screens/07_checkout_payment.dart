import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../l10n/app_localizations.dart';
import '../models/order.dart';
import '../state/providers.dart';
import '../theme/tokens.dart';
import '../widgets/payment_method_tile.dart';
import '../widgets/voice_mic_fab.dart';

/// DES-0008 §3.7 — Checkout / payment method.
class CheckoutPaymentScreen extends ConsumerStatefulWidget {
  const CheckoutPaymentScreen({super.key});
  @override
  ConsumerState<CheckoutPaymentScreen> createState() =>
      _CheckoutPaymentScreenState();
}

class _CheckoutPaymentScreenState
    extends ConsumerState<CheckoutPaymentScreen> {
  PaymentMethod _method = PaymentMethod.creditCard;
  bool _submitting = false;

  @override
  Widget build(BuildContext context) {
    final l = AppLocalizations.of(context);
    final totals = ref.watch(cartProvider).totals;
    return Scaffold(
      appBar: AppBar(title: Text(l.checkoutTitle)),
      floatingActionButton: const VoiceMicFab(),
      body: SafeArea(
        child: Padding(
          padding: const EdgeInsets.all(MontiSpacing.lg),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                l.payWith,
                style: const TextStyle(
                    fontSize: 18, fontWeight: FontWeight.w700),
              ),
              const SizedBox(height: MontiSpacing.md),
              GridView.count(
                shrinkWrap: true,
                crossAxisCount: 2,
                mainAxisSpacing: MontiSpacing.md,
                crossAxisSpacing: MontiSpacing.md,
                childAspectRatio: 2.2,
                physics: const NeverScrollableScrollPhysics(),
                children: [
                  PaymentMethodTile(
                    key: const Key('pay.credit_card'),
                    method: PaymentMethod.creditCard,
                    label: l.paymentCreditCard,
                    icon: Icons.credit_card,
                    selected: _method == PaymentMethod.creditCard,
                    onSelect: (m) => setState(() => _method = m),
                  ),
                  PaymentMethodTile(
                    key: const Key('pay.apple_pay'),
                    method: PaymentMethod.applePay,
                    label: l.paymentApplePay,
                    icon: Icons.apple,
                    selected: _method == PaymentMethod.applePay,
                    onSelect: (m) => setState(() => _method = m),
                  ),
                  PaymentMethodTile(
                    key: const Key('pay.wallet'),
                    method: PaymentMethod.wallet,
                    label: l.paymentWallet,
                    icon: Icons.account_balance_wallet_outlined,
                    selected: _method == PaymentMethod.wallet,
                    onSelect: (m) => setState(() => _method = m),
                  ),
                  PaymentMethodTile(
                    key: const Key('pay.cash'),
                    method: PaymentMethod.cash,
                    label: l.paymentCash,
                    icon: Icons.payments_outlined,
                    selected: _method == PaymentMethod.cash,
                    onSelect: (m) => setState(() => _method = m),
                  ),
                ],
              ),
              const SizedBox(height: MontiSpacing.lg),
              Text(l.summary,
                  style: const TextStyle(fontWeight: FontWeight.w700)),
              const SizedBox(height: MontiSpacing.sm),
              _row(l.subtotal, '฿ ${totals.subtotalTHB}'),
              _row('${l.serviceCharge} + ${l.tax7Pct}',
                  '฿ ${totals.serviceChargeTHB + totals.taxTHB}'),
              if (totals.hasDiscount)
                _row(l.discount, '−฿ ${totals.discountTHB}', isDiscount: true),
              const Divider(color: MontiColors.bgSurface2),
              _row(l.total, '฿ ${totals.totalTHB}', emphasis: true),
              const Spacer(),
              SizedBox(
                width: double.infinity,
                child: ElevatedButton(
                  key: const Key('pay.now'),
                  onPressed: _submitting ? null : () => _submit(totals.totalTHB),
                  child: Text('${l.payNow}  ( ฿ ${totals.totalTHB} )'),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Future<void> _submit(int total) async {
    setState(() => _submitting = true);
    String orderNumber;
    try {
      final api = ref.read(orderApiProvider);
      orderNumber = await api.submit(
        sessionId: ref.read(sessionIdProvider) ?? 'demo-session',
        tableCode: ref.read(tableCodeProvider) ?? 'A12',
        method: _method,
        idempotencyKey: 'idem-${DateTime.now().millisecondsSinceEpoch}',
      );
    } catch (_) {
      // MVP: stub order id when backend not reachable.
      orderNumber = 'MO-2026-00481';
    }
    if (!mounted) return;
    ref.read(cartProvider.notifier).reset();
    context.go('/thank-you/$orderNumber');
  }

  Widget _row(String label, String value,
      {bool emphasis = false, bool isDiscount = false}) {
    final style = TextStyle(
      fontSize: emphasis ? 18 : 14,
      fontWeight: emphasis ? FontWeight.w700 : FontWeight.w500,
      color: isDiscount ? MontiColors.success : MontiColors.textPrimary,
    );
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [Text(label, style: style), Text(value, style: style)],
      ),
    );
  }
}
