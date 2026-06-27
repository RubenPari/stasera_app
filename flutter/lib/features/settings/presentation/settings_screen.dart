import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'sections/account_section.dart';
import 'sections/preferences_section.dart';
import 'sections/staples_section.dart';

/// Schermata impostazioni: preferenze, staple, account.
///
/// God-class scremata: ogni sezione è un widget proprio sotto `sections/`.
class SettingsScreen extends ConsumerWidget {
  const SettingsScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    return Scaffold(
      appBar: AppBar(title: const Text('Impostazioni')),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          _sectionTitle(context, 'Le mie preferenze'),
          const PreferencesSection(),
          const SizedBox(height: 24),
          _sectionTitle(context, 'Prodotti sempre in casa'),
          const StaplesSection(),
          const SizedBox(height: 24),
          _sectionTitle(context, 'Account'),
          const AccountSection(),
        ],
      ),
    );
  }

  Widget _sectionTitle(BuildContext context, String title) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 12),
      child: Text(
        title,
        style: const TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
      ),
    );
  }
}
