import 'dart:convert';

import '../models/menu.dart';
import 'api_client.dart';

/// Wraps `GET /api/v1/menu` from TASK-0009.
class MenuApi {
  MenuApi(this._client);
  final ApiClient _client;

  Future<List<MenuCategory>> fetchCategories() async {
    final r = await _client.get('/api/v1/menu/categories');
    if (r.statusCode != 200) {
      throw ApiError('menu/categories', r.statusCode, r.body);
    }
    final list = (jsonDecode(r.body) as Map<String, dynamic>)['categories']
        as List<dynamic>;
    return list
        .map((j) => MenuCategory.fromJson(j as Map<String, dynamic>))
        .toList(growable: false);
  }

  Future<List<MenuItem>> fetchItems({String? categoryId}) async {
    final r = await _client.get(
      '/api/v1/menu/items',
      query: categoryId == null ? null : {'category_id': categoryId},
    );
    if (r.statusCode != 200) {
      throw ApiError('menu/items', r.statusCode, r.body);
    }
    final list =
        (jsonDecode(r.body) as Map<String, dynamic>)['items'] as List<dynamic>;
    return list
        .map((j) => MenuItem.fromJson(j as Map<String, dynamic>))
        .toList(growable: false);
  }

  Future<MenuItem> fetchItem(String id) async {
    final r = await _client.get('/api/v1/menu/items/$id');
    if (r.statusCode != 200) {
      throw ApiError('menu/items/$id', r.statusCode, r.body);
    }
    return MenuItem.fromJson(jsonDecode(r.body) as Map<String, dynamic>);
  }
}

class ApiError implements Exception {
  ApiError(this.endpoint, this.statusCode, this.body);
  final String endpoint;
  final int statusCode;
  final String body;
  @override
  String toString() => 'ApiError($endpoint, $statusCode): $body';
}
