import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/models/dto.dart';
import '../../../core/theme/app_theme.dart';
import '../../../shared/util/load_on_post_frame.dart';
import '../../../shared/widgets/loading_overlay.dart';
import '../providers/shopping_provider.dart';

/// Lista della spesa raggruppata per corsia, con checkbox e barra progresso.
class ShoppingScreen extends ConsumerStatefulWidget {
  const ShoppingScreen({super.key});

  @override
  ConsumerState<ShoppingScreen> createState() => _ShoppingScreenState();
}

class _ShoppingScreenState extends ConsumerState<ShoppingScreen> {
  @override
  void initState() {
    super.initState();
    loadOnPostFrame(() => ref.read(shoppingProvider.notifier).load());
  }

  @override
  Widget build(BuildContext context) {
    final state = ref.watch(shoppingProvider);

    return Scaffold(
      appBar: AppBar(title: const Text('Spesa')),
      body: Stack(
        children: [
          RefreshIndicator(
            onRefresh: () => ref.read(shoppingProvider.notifier).load(),
            child: _body(state),
          ),
          LoadingOverlay(
            message: 'Genero la lista della spesa...',
            visible: state is ShoppingLoading,
          ),
        ],
      ),
    );
  }

  Widget _body(ShoppingState state) {
    if (state is ShoppingInitial || state is ShoppingLoading) {
      return const Center(child: CircularProgressIndicator());
    }
    if (state is ShoppingError) {
      return Center(child: Text('Errore: ${state.message}'));
    }
    if (state is ShoppingEmpty) {
      return Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            const Icon(Icons.shopping_cart_outlined, size: 64),
            const SizedBox(height: 16),
            const Text('Nessuna lista della spesa attiva.'),
            const SizedBox(height: 16),
            FilledButton.icon(
              onPressed: () => ref.read(shoppingProvider.notifier).generate(),
              icon: const Icon(Icons.auto_awesome),
              label: const Text('Genera dal piano'),
            ),
          ],
        ),
      );
    }
    if (state is ShoppingLoaded) {
      final grouped = _groupByAisle(state.list.items);
      final progress = state.list.items.isEmpty
          ? 0.0
          : state.checkedCount / state.list.items.length;
      final isCompleted = state.list.completedAt != null;

      return Column(
        children: [
          // Barra progresso.
          Padding(
            padding: const EdgeInsets.all(16),
            child: Column(
              children: [
                LinearProgressIndicator(value: progress),
                const SizedBox(height: 8),
                Text(
                    '${state.checkedCount}/${state.list.items.length} spuntati',),
              ],
            ),
          ),
          Expanded(
            child: ListView(
              padding: const EdgeInsets.symmetric(horizontal: 16),
              children: [
                ...grouped.entries
                    .map((e) => _aisleSection(e.key, e.value, isCompleted)),
                const SizedBox(height: 24),
                if (state.allChecked && !isCompleted)
                  FilledButton.icon(
                    onPressed: () =>
                        ref.read(shoppingProvider.notifier).complete(),
                    icon: const Icon(Icons.check),
                    label: const Text('Spesa completata'),
                  ),
                if (isCompleted)
                  const Padding(
                    padding: EdgeInsets.all(16),
                    child: Text(
                      'Spesa completata ✓',
                      textAlign: TextAlign.center,
                      style: TextStyle(color: AppTheme.success, fontSize: 18),
                    ),
                  ),
              ],
            ),
          ),
        ],
      );
    }
    return const SizedBox.shrink();
  }

  Widget _aisleSection(
      String aisle, List<ShoppingItemDto> items, bool isCompleted,) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Padding(
          padding: const EdgeInsets.symmetric(vertical: 12),
          child: Text(
            _aisleLabel(aisle),
            style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 16),
          ),
        ),
        ...items.map((i) => Card(
              child: CheckboxListTile(
                value: i.isChecked,
                onChanged: isCompleted
                    ? null
                    : (v) => ref
                        .read(shoppingProvider.notifier)
                        .toggleItem(i.id, v ?? false),
                title: Text(
                  i.name,
                  style: TextStyle(
                    decoration: i.isChecked ? TextDecoration.lineThrough : null,
                  ),
                ),
                subtitle: Text(i.quantity),
              ),
            ),),
      ],
    );
  }

  /// Raggruppa gli item per corsia con un ordinamento fisso (carne, frigo,
  /// verdura, dispensa, altro) e tipizzato con [ShoppingItemDto].
  Map<String, List<ShoppingItemDto>> _groupByAisle(
      List<ShoppingItemDto> items,) {
    final m = <String, List<ShoppingItemDto>>{};
    for (final i in items) {
      m.putIfAbsent(i.aisle, () => []).add(i);
    }
    const order = ['carne', 'frigo', 'verdura', 'dispensa', 'altro'];
    final sorted = <String, List<ShoppingItemDto>>{};
    for (final a in order) {
      if (m.containsKey(a)) sorted[a] = m[a]!;
    }
    sorted.addAll(m..removeWhere((k, _) => order.contains(k)));
    return sorted;
  }

  String _aisleLabel(String aisle) {
    const labels = {
      'carne': 'Carne e pesce',
      'frigo': 'Frigo',
      'verdura': 'Frutta e verdura',
      'dispensa': 'Dispensa',
      'altro': 'Altro',
    };
    return labels[aisle] ?? aisle;
  }
}
