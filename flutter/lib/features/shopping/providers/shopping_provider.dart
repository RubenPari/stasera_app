import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/models/dto.dart';
import '../data/shopping_repository.dart';

sealed class ShoppingState {
  const ShoppingState();
}

class ShoppingInitial extends ShoppingState {
  const ShoppingInitial();
}

class ShoppingLoading extends ShoppingState {
  const ShoppingLoading();
}

class ShoppingEmpty extends ShoppingState {
  const ShoppingEmpty();
}

class ShoppingLoaded extends ShoppingState {
  const ShoppingLoaded(this.list);
  final ShoppingListDto list;
  int get checkedCount => list.items.where((i) => i.isChecked).length;
  bool get allChecked =>
      list.items.isNotEmpty && checkedCount == list.items.length;
}

class ShoppingError extends ShoppingState {
  const ShoppingError(this.message);
  final String message;
}

class ShoppingNotifier extends StateNotifier<ShoppingState> {
  ShoppingNotifier(this._repo) : super(const ShoppingInitial());

  final ShoppingRepository _repo;

  Future<void> load() async {
    state = const ShoppingLoading();
    try {
      final list = await _repo.getCurrent();
      if (list == null) {
        state = const ShoppingEmpty();
      } else {
        state = ShoppingLoaded(list);
      }
    } catch (e) {
      state = ShoppingError(e.toString());
    }
  }

  Future<void> generate() async {
    state = const ShoppingLoading();
    try {
      final list = await _repo.generate();
      state = ShoppingLoaded(list);
    } catch (e) {
      state = ShoppingError(e.toString());
    }
  }

  Future<void> toggleItem(String itemId, bool isChecked) async {
    final s = state;
    if (s is! ShoppingLoaded) return;
    try {
      final updated = await _repo.updateItem(itemId, isChecked);
      final items =
          s.list.items.map((i) => i.id == itemId ? updated : i).toList();
      state = ShoppingLoaded(s.list.copyWith(items: items));
    } catch (e) {
      state = ShoppingError(e.toString());
    }
  }

  Future<void> complete() async {
    final s = state;
    if (s is! ShoppingLoaded) return;
    try {
      final list = await _repo.complete();
      state = ShoppingLoaded(list);
    } catch (e) {
      state = ShoppingError(e.toString());
    }
  }
}

final shoppingProvider =
    StateNotifierProvider.autoDispose<ShoppingNotifier, ShoppingState>((ref) {
  return ShoppingNotifier(ref.watch(shoppingRepositoryProvider));
});
