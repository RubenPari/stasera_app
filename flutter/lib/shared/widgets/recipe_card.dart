import 'package:flutter/material.dart';

import '../../../core/models/dto.dart';

/// Card grande che mostra la ricetta della cena.
class RecipeCard extends StatelessWidget {
  const RecipeCard({super.key, required this.recipe, this.onTap});

  final RecipeDto recipe;
  final VoidCallback? onTap;

  @override
  Widget build(BuildContext context) {
    return Card(
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(16),
        child: Padding(
          padding: const EdgeInsets.all(20),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                children: [
                  const Icon(Icons.restaurant_menu, size: 28),
                  const SizedBox(width: 8),
                  Expanded(
                    child: Text(
                      recipe.name,
                      style: Theme.of(context).textTheme.headlineSmall,
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 16),
              Row(
                children: [
                  _MetricChip(
                    icon: Icons.timer_outlined,
                    label: '${recipe.prepMinutes} min',
                  ),
                  const SizedBox(width: 12),
                  _MetricChip(
                    icon: Icons.kitchen,
                    label: '${recipe.ingredients.length} ingredienti',
                  ),
                  if (recipe.isRescue) ...[
                    const SizedBox(width: 12),
                    _MetricChip(
                      icon: Icons.bolt,
                      label: 'emergenza',
                      color: Theme.of(context).colorScheme.secondary,
                    ),
                  ],
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _MetricChip extends StatelessWidget {
  const _MetricChip({
    required this.icon,
    required this.label,
    this.color,
  });

  final IconData icon;
  final String label;
  final Color? color;

  @override
  Widget build(BuildContext context) {
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        Icon(icon, size: 18, color: color),
        const SizedBox(width: 4),
        Text(label, style: TextStyle(color: color)),
      ],
    );
  }
}
