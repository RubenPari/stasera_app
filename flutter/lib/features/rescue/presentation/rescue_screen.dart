import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../shared/widgets/recipe_card.dart';
import '../providers/rescue_provider.dart';

/// Modalità "troppo stanco": 3 ricette di emergenza con staple.
class RescueScreen extends ConsumerStatefulWidget {
  const RescueScreen({super.key});

  @override
  ConsumerState<RescueScreen> createState() => _RescueScreenState();
}

class _RescueScreenState extends ConsumerState<RescueScreen> {
  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      ref.read(rescueProvider.notifier).load();
    });
  }

  @override
  Widget build(BuildContext context) {
    final state = ref.watch(rescueProvider);

    return Scaffold(
      appBar: AppBar(
        title: const Text('Stasera cucino poco'),
        leading: IconButton(
          icon: const Icon(Icons.arrow_back),
          onPressed: () => context.go('/'),
        ),
      ),
      body: state is RescueInitial || state is RescueLoading
          ? Center(
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: const [
                  CircularProgressIndicator(),
                  SizedBox(height: 16),
                  Text('Sto cercando cosa hai in casa...'),
                ],
              ),
            )
          : state is RescueError
              ? Center(child: Text('Errore: ${state.message}'))
              : state is RescueLoaded
                  ? ListView(
                      padding: const EdgeInsets.all(16),
                      children: [
                        const Text(
                          'Tre opzioni in 10 minuti o meno:',
                          style: TextStyle(fontWeight: FontWeight.bold),
                        ),
                        const SizedBox(height: 12),
                        ...state.recipes.map((r) => Padding(
                              padding: const EdgeInsets.only(bottom: 12),
                              child: RecipeCard(
                                recipe: r,
                                onTap: () => context.push('/cooking/${r.id}'),
                              ),
                            )),
                      ],
                    )
                  : const SizedBox.shrink(),
    );
  }
}