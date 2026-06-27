import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../providers/settings_provider.dart';

/// Sezione "Prodotti sempre in casa": lista staple con toggle attivo e delete,
/// più aggiunta di nuovi staple.
///
/// Estratta dalla god-class `SettingsScreen`.
class StaplesSection extends ConsumerStatefulWidget {
  const StaplesSection({super.key});

  @override
  ConsumerState<StaplesSection> createState() => _StaplesSectionState();
}

class _StaplesSectionState extends ConsumerState<StaplesSection> {
  final _newStapleCtrl = TextEditingController();

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback(
      (_) => ref.read(staplesProvider.notifier).load(),
    );
  }

  @override
  void dispose() {
    _newStapleCtrl.dispose();
    super.dispose();
  }

  Future<void> _addStaple() async {
    final name = _newStapleCtrl.text.trim();
    if (name.isEmpty) return;
    _newStapleCtrl.clear();
    await ref.read(staplesProvider.notifier).createStaple(name);
  }

  @override
  Widget build(BuildContext context) {
    final state = ref.watch(staplesProvider);

    if (state is StaplesLoading || state is StaplesInitial) {
      return const Center(child: CircularProgressIndicator());
    }
    if (state is StaplesError) {
      return Text('Errore: ${state.message}');
    }
    final list = (state as StaplesLoaded).staples;
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            ...list.map((s) => SwitchListTile(
                  value: s.isActive,
                  title: Text(s.name),
                  onChanged: (v) =>
                      ref.read(staplesProvider.notifier).updateStaple(s.id, v),
                  secondary: IconButton(
                    icon: const Icon(Icons.delete_outline),
                    onPressed: () =>
                        ref.read(staplesProvider.notifier).deleteStaple(s.id),
                  ),
                ),),
            const Divider(),
            Row(
              children: [
                Expanded(
                  child: TextField(
                    controller: _newStapleCtrl,
                    decoration:
                        const InputDecoration(hintText: 'Aggiungi staple'),
                  ),
                ),
                IconButton(
                  icon: const Icon(Icons.add),
                  onPressed: _addStaple,
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }
}
