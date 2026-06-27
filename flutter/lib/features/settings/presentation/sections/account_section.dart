import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../auth/providers/auth_provider.dart';

/// Sezione "Account": nome (modificabile), email, cambio password, logout.
///
/// Estratta dalla god-class `SettingsScreen`. Il logout naviga a `/login`
/// tramite `AuthNotifier.logout` (wiring GoRouter condiviso con
/// `AuthInterceptor`).
class AccountSection extends ConsumerStatefulWidget {
  const AccountSection({super.key});

  @override
  ConsumerState<AccountSection> createState() => _AccountSectionState();
}

class _AccountSectionState extends ConsumerState<AccountSection> {
  final _editNameCtrl = TextEditingController();
  final _currentPwdCtrl = TextEditingController();
  final _newPwdCtrl = TextEditingController();
  bool _savingName = false;
  bool _savingPassword = false;

  @override
  void dispose() {
    _editNameCtrl.dispose();
    _currentPwdCtrl.dispose();
    _newPwdCtrl.dispose();
    super.dispose();
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
                onSubmitted: (_) =>
                    Navigator.of(ctx).pop(_editNameCtrl.text.trim()),
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
    if (result == null || result.isEmpty || result == current || !mounted) {
      return;
    }
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
                    decoration:
                        const InputDecoration(labelText: 'Password attuale'),
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
                  onPressed: _savingPassword
                      ? null
                      : () => Navigator.of(ctx).pop(false),
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
      messenger
          .showSnackBar(const SnackBar(content: Text('Password aggiornata')));
    } else {
      messenger.showSnackBar(SnackBar(content: Text('Errore: $err')));
    }
  }

  @override
  Widget build(BuildContext context) {
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
              onTap: user == null
                  ? null
                  : () => _editDisplayName(user.displayName),
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
                onPressed: () => ref.read(authProvider.notifier).logout(),
                icon: const Icon(Icons.logout),
                label: const Text('Esci'),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
