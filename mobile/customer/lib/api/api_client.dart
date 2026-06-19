import 'package:http/http.dart' as http;

/// Thin HTTP client wrapper. The backend base URL is read from
/// `--dart-define=MONTI_API_BASE=https://...` at build time; falls back to
/// `http://localhost:8080` for `flutter run` against compose.e2e.yml.
class ApiClient {
  ApiClient({http.Client? inner, String? baseUrl})
      : _inner = inner ?? http.Client(),
        baseUrl = baseUrl ?? _defaultBase();

  final http.Client _inner;
  final String baseUrl;

  static String _defaultBase() => const String.fromEnvironment(
        'MONTI_API_BASE',
        defaultValue: 'http://localhost:8080',
      );

  Uri uri(String path, [Map<String, String>? query]) =>
      Uri.parse(baseUrl).replace(
        path: path,
        queryParameters: query,
      );

  Future<http.Response> get(String path, {Map<String, String>? query}) =>
      _inner.get(uri(path, query));

  Future<http.Response> post(String path, {Object? body}) => _inner.post(
        uri(path),
        headers: const {'content-type': 'application/json'},
        body: body,
      );

  void close() => _inner.close();
}
