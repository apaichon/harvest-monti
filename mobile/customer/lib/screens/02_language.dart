import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../l10n/app_localizations.dart';
import '../state/providers.dart';
import '../theme/tokens.dart';
import '../widgets/voice_mic_fab.dart';

const _languages = <({String code, String flag, String label})>[
  (code: 'en', flag: 'GB', label: 'English'),
  (code: 'th', flag: 'TH', label: 'ภาษาไทย'),
  (code: 'zh', flag: 'CN', label: '中文'),
  (code: 'ja', flag: 'JP', label: '日本語'),
];

/// DES-0008 §3.2 — Language selection with persistence (REQ-0010 AC-2/AC-11).
class LanguageSelectionScreen extends ConsumerStatefulWidget {
  const LanguageSelectionScreen({super.key});

  @override
  ConsumerState<LanguageSelectionScreen> createState() =>
      _LanguageSelectionScreenState();
}

class _LanguageSelectionScreenState
    extends ConsumerState<LanguageSelectionScreen> {
  String? _selected;

  @override
  void initState() {
    super.initState();
    _selected = ref.read(localeProvider).languageCode;
  }

  @override
  Widget build(BuildContext context) {
    final l = AppLocalizations.of(context);
    return Scaffold(
      appBar: AppBar(title: Text(l.languageTitle)),
      floatingActionButton: const VoiceMicFab(),
      body: SafeArea(
        child: Padding(
          padding: const EdgeInsets.all(MontiSpacing.lg),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                l.chooseLanguage,
                style: const TextStyle(
                  fontSize: 22,
                  fontWeight: FontWeight.w700,
                ),
              ),
              const SizedBox(height: MontiSpacing.lg),
              Expanded(
                child: GridView.count(
                  crossAxisCount: 2,
                  mainAxisSpacing: MontiSpacing.md,
                  crossAxisSpacing: MontiSpacing.md,
                  childAspectRatio: 1.1,
                  children: [
                    for (final lang in _languages) _card(lang),
                  ],
                ),
              ),
              SizedBox(
                width: double.infinity,
                child: ElevatedButton(
                  key: const Key('language.next'),
                  onPressed: _selected == null
                      ? null
                      : () async {
                          await ref
                              .read(localeProvider.notifier)
                              .set(Locale(_selected!));
                          if (mounted) context.go('/categories');
                        },
                  child: Text(l.next),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _card(({String code, String flag, String label}) lang) {
    final selected = _selected == lang.code;
    return InkWell(
      key: Key('language.card.${lang.code}'),
      borderRadius: BorderRadius.circular(MontiRadii.card),
      onTap: () => setState(() => _selected = lang.code),
      child: Container(
        padding: const EdgeInsets.all(MontiSpacing.md),
        decoration: BoxDecoration(
          color: selected ? MontiColors.bgSurface2 : MontiColors.bgSurface,
          borderRadius: BorderRadius.circular(MontiRadii.card),
          border: Border.all(
            color: selected ? MontiColors.accentCyan : Colors.transparent,
            width: 2,
          ),
        ),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Text(
              lang.flag,
              style: const TextStyle(fontSize: 32, fontWeight: FontWeight.w700),
            ),
            const SizedBox(height: MontiSpacing.sm),
            Text(
              lang.label,
              style: const TextStyle(fontSize: 16, fontWeight: FontWeight.w600),
            ),
            const SizedBox(height: MontiSpacing.sm),
            Radio<String>(
              value: lang.code,
              groupValue: _selected,
              onChanged: (v) => setState(() => _selected = v),
              activeColor: MontiColors.accentCyan,
            ),
          ],
        ),
      ),
    );
  }
}
