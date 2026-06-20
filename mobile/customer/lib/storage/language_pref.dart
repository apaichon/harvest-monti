import 'package:shared_preferences/shared_preferences.dart';

/// Persisted language selection. Survives app restart per REQ-0010 AC-11.
class LanguagePref {
  LanguagePref(this._prefs);
  final SharedPreferences _prefs;
  static const _key = 'monti.locale';

  static Future<LanguagePref> open() async {
    final p = await SharedPreferences.getInstance();
    return LanguagePref(p);
  }

  String? read() => _prefs.getString(_key);

  Future<void> write(String code) => _prefs.setString(_key, code);
}
