import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/models/dto.dart';
import '../../auth/providers/auth_provider.dart';
import '../data/settings_repository.dart';
import '../providers/settings_provider.dart';

/// Schermata impostazioni: preferenze, staple, account.
class SettingsScreen extends ConsumerStatefulWidget {
  const SettingsScreen({super.key});

  @override
  ConsumerState<SettingsScreen> createState() => _SettingsScreenState();
}

class _SettingsScreenState extends ConsumerState<SettingsScreen> {
  final _newStapleCtrl = TextEditingController();
  final _newDislikedCtrl = TextEditingController();
  final _newCuisineCtrl = TextEditingController();
  final _editNameCtrl = TextEditingController();
  final _currentPwdCtrl = TextEditingController();
  final _newPwdCtrl = TextEditingController();
  int? _pendingMaxPrep;
  bool _savingName = false;
  bool _savingPassword = false;

  @override
  void dispose() {
    _newStapleCtrl.dispose();
    _newDislikedCtrl.dispose();
    _newCuisineCtrl.dispose();
    _editNameCtrl.dispose();
    _currentPwdCtrl.dispose();
    _newPwdCtrl.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final prefs = ref.watch(preferencesProvider);
    final staples = ref.watch(staplesProvider);

    return Scaffold(
      appBar: AppBar(title: const Text('Impostazioni')),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          _sectionTitle('Le mie preferenze'),
          prefs.when(
            data: (p) => _preferencesSection(p),
            loading: () => const Center(child: CircularProgressIndicator()),
            error: (e, _) => Text('Errore: $e'),
          ),
          const SizedBox(height: 24),
          _sectionTitle('Prodotti sempre in casa'),
          staples.when(
            data: (list) => _staplesSection(list),
            loading: () => const Center(child: CircularProgressIndicator()),
            error: (e, _) => Text('Errore: $e'),
          ),
          const SizedBox(height: 24),
          _sectionTitle('Account'),
          _accountSection(),
        ],
      ),
    );
  }

  Widget _sectionTitle(String title) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 12),
      child: Text(
        title,
        style: const TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
      ),
    );
  }

  Widget _preferencesSection(PreferencesDto p) {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Text('Cibi da evitare'),
            Wrap(
              spacing: 8,
              children: p.dislikedIngredients.map((d) => Chip(
                label: Text(d),
                onDeleted: () => _removeDisliked(p, d),
              )).toList(),
            ),
            const SizedBox(height: 8),
            Row(
              children: [
                Expanded(
                  child: TextField(
                    controller: _newDislikedCtrl,
                    decoration: const InputDecoration(hintText: 'Aggiungi cibo da evitare'),
                  ),
                ),
                IconButton(
                  icon: const Icon(Icons.add),
                  onPressed: () => _addDisliked(p),
                ),
              ],
            ),
            const SizedBox(height: 16),
            const Text('Cucine preferite'),
            Wrap(
              spacing: 8,
              children: p.preferredCuisines.map((c) => Chip(
                label: Text(c),
                onDeleted: () => _removeCuisine(p, c),
              )).toList(),
            ),
            const SizedBox(height: 8),
            Row(
              children: [
                Expanded(
                  child: TextField(
                    controller: _newCuisineCtrl,
                    decoration: const InputDecoration(hintText: 'Aggiungi cucina'),
                  ),
                ),
                IconButton(
                  icon: const Icon(Icons.add),
                  onPressed: () => _addCuisine(p),
                ),
              ],
            ),
            const SizedBox(height: 16),
            Text('Tempo massimo: ${_pendingMaxPrep ?? p.maxPrepMinutes} min'),
            Slider(
              min: 5,
              max: 120,
              divisions: 23,
              value: ((_pendingMaxPrep ?? p.maxPrepMinutes).clamp(5, 120)).toDouble(),
              label: (_pendingMaxPrep ?? p.maxPrepMinutes).toString(),
              onChanged: (v) => setState(() => _pendingMaxPrep = v.round()),
              onChangeEnd: (v) async {
                final next = v.round();
                try {
                  await ref.read(settingsRepositoryProvider).updatePreferences(
                        disliked: p.dislikedIngredients,
                        maxPrep: next,
                        cuisines: p.preferredCuisines,
                      );
                  ref.invalidate(preferencesProvider);
                  if (mounted) {
                    setState(() => _pendingMaxPrep = null);
                  }
                } catch (e) {
                  if (mounted) {
                    ScaffoldMessenger.of(context).showSnackBar(
                      SnackBar(content: Text('Errore: $e')),
                    );
                  }
                }
              },
            ),
          ],
        ),
      ),
    );
  }

  Widget _staplesSection(List<StapleDto> list) {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            ...list.map((s) => SwitchListTile(
                  value: s.isActive,
                  title: Text(s.name),
                  onChanged: (v) async {
                    await ref.read(settingsRepositoryProvider).updateStaple(s.id, v);
                    ref.invalidate(staplesProvider);
                  },
                  secondary: IconButton(
                    icon: const Icon(Icons.delete_outline),
                    onPressed: () async {
                      await ref.read(settingsRepositoryProvider).deleteStaple(s.id);
                      ref.invalidate(staplesProvider);
                    },
                  ),
                )),
            const Divider(),
            Row(
              children: [
                Expanded(
                  child: TextField(
                    controller: _newStapleCtrl,
                    decoration: const InputDecoration(hintText: 'Aggiungi staple'),
                  ),
                ),
                IconButton(
                  icon: const Icon(Icons.add),
                  onPressed: () async {
                    if (_newStapleCtrl.text.trim().isEmpty) return;
                    await ref
                        .read(settingsRepositoryProvider)
                        .createStaple(_newStapleCtrl.text.trim());
                    _newStapleCtrl.clear();
                    ref.invalidate(staplesProvider);
                  },
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }

  Widget _accountSection() {
    final authState = ref.watch(authProvider);
    final user = authState is Authenticated ? authState.user : null;

    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            ListTile(
              contentPadding: EdgeInsets.zero,
              leading: const Icon(Icons.person),
              title: const Text('Nome'),
              subtitle: Text(user?.displayName ?? '—'),
              trailing: const Icon(Icons.edit_outlined),
              onTap: user == null ? null : () => _editDisplayName(user.displayName),
            ),
            ListTile(
              contentPadding: EdgeInsets.zero,
              leading: const Icon(Icons.email_outlined),
              title: const Text('Email'),
              subtitle: Text(user?.email ?? '—'),
            ),
            ListTile(
              contentPadding: EdgeInsets.zero,
              leading: const Icon(Icons.lock_outline),
              title: const Text('Password'),
              trailing: TextButton(
                onPressed: _savingPassword ? null : _changePassword,
                child: const Text('Cambia password'),
              ),
            ),
            const Divider(),
            Align(
              alignment: Alignment.centerLeft,
              child: FilledButton.tonalIcon(
                onPressed: () async {
                  await ref.read(authProvider.notifier).logout();
                },
                icon: const Icon(Icons.logout),
                label: const Text('Esci'),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Future<void> _editDisplayName(String current) async {
    _editNameCtrl.text = current;
    final result = await showDialog<String?>(
      context: context,
      builder: (ctx) {
        return StatefulBuilder(
          builder: (ctx, setLocal) {
            return AlertDialog(
              title: const Text('Modifica nome'),
              content: TextField(
                controller: _editNameCtrl,
                autofocus: true,
                decoration: const InputDecoration(labelText: 'Nome'),
                textInputAction: TextInputAction.done,
                onSubmitted: (_) => Navigator.of(ctx).pop(_editNameCtrl.text.trim()),
              ),
              actions: [
                TextButton(
                  onPressed: _savingName ? null : () => Navigator.of(ctx).pop(),
                  child: const Text('Annulla'),
                ),
                FilledButton(
                  onPressed: _savingName
                      ? null
                      : () => Navigator.of(ctx).pop(_editNameCtrl.text.trim()),
                  child: _savingName
                      ? const SizedBox(
                          width: 16,
                          height: 16,
                          child: CircularProgressIndicator(strokeWidth: 2),
                        )
                      : const Text('OK'),
                ),
              ],
            );
          },
        );
      },
    );
    if (result == null || result.isEmpty || result == current || !mounted) return;
    setState(() => _savingName = true);
    final err = await ref.read(authProvider.notifier).updateDisplayName(result);
    if (!mounted) return;
    setState(() => _savingName = false);
    final messenger = ScaffoldMessenger.of(context);
    if (err == null) {
      messenger.showSnackBar(const SnackBar(content: Text('Salvato')));
    } else {
      messenger.showSnackBar(SnackBar(content: Text('Errore: $err')));
    }
  }

  Future<void> _changePassword() async {
    _currentPwdCtrl.clear();
    _newPwdCtrl.clear();
    final result = await showDialog<bool>(
      context: context,
      builder: (ctx) {
        return StatefulBuilder(
          builder: (ctx, setLocal) {
            return AlertDialog(
              title: const Text('Cambia password'),
              content: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  TextField(
                    controller: _currentPwdCtrl,
                    obscureText: true,
                    decoration: const InputDecoration(labelText: 'Password attuale'),
                  ),
                  const SizedBox(height: 12),
                  TextField(
                    controller: _newPwdCtrl,
                    obscureText: true,
                    decoration: const InputDecoration(
                      labelText: 'Nuova password (≥ 8 caratteri)',
                    ),
                  ),
                ],
              ),
              actions: [
                TextButton(
                  onPressed: _savingPassword ? null : () => Navigator.of(ctx).pop(false),
                  child: const Text('Annulla'),
                ),
                FilledButton(
                  onPressed: _savingPassword || _newPwdCtrl.text.length < 8
                      ? null
                      : () => Navigator.of(ctx).pop(true),
                  child: _savingPassword
                      ? const SizedBox(
                          width: 16,
                          height: 16,
                          child: CircularProgressIndicator(strokeWidth: 2),
                        )
                      : const Text('OK'),
                ),
              ],
            );
          },
        );
      },
    );
    if (result != true || !mounted) return;
    setState(() => _savingPassword = true);
    final err = await ref.read(authProvider.notifier).changePassword(
          _currentPwdCtrl.text,
          _newPwdCtrl.text,
        );
    if (!mounted) return;
    setState(() => _savingPassword = false);
    final messenger = ScaffoldMessenger.of(context);
    if (err == null) {
      messenger.showSnackBar(const SnackBar(content: Text('Password aggiornata')));
    } else {
      messenger.showSnackBar(SnackBar(content: Text('Errore: $err')));
    }
  }

  Future<void> _addDisliked(PreferencesDto p) async {
    final name = _newDislikedCtrl.text.trim();
    if (name.isEmpty) return;
    final list = [...p.dislikedIngredients, name];
    await ref.read(settingsRepositoryProvider).updatePreferences(
          disliked: list,
          maxPrep: p.maxPrepMinutes,
          cuisines: p.preferredCuisines,
        );
    _newDislikedCtrl.clear();
    ref.invalidate(preferencesProvider);
  }

  Future<void> _removeDisliked(PreferencesDto p, String name) async {
    final list = p.dislikedIngredients.where((e) => e != name).toList();
    await ref.read(settingsRepositoryProvider).updatePreferences(
          disliked: list,
          maxPrep: p.maxPrepMinutes,
          cuisines: p.preferredCuisines,
        );
    ref.invalidate(preferencesProvider);
  }

  Future<void> _addCuisine(PreferencesDto p) async {
    final name = _newCuisineCtrl.text.trim();
    if (name.isEmpty) return;
    final list = [...p.preferredCuisines, name];
    await ref.read(settingsRepositoryProvider).updatePreferences(
          disliked: p.dislikedIngredients,
          maxPrep: p.maxPrepMinutes,
          cuisines: list,
        );
    _newCuisineCtrl.clear();
    ref.invalidate(preferencesProvider);
  }

  Future<void> _removeCuisine(PreferencesDto p, String name) async {
    final list = p.preferredCuisines.where((e) => e != name).toList();
    await ref.read(settingsRepositoryProvider).updatePreferences(
          disliked: p.dislikedIngredients,
          maxPrep: p.maxPrepMinutes,
          cuisines: list,
        );
    ref.invalidate(preferencesProvider);
  }
}