import 'package:flutter/material.dart';

import '../../../core/models/dto.dart';

/// Card che rappresenta uno step di cottura con eventuale timer.
class StepCard extends StatelessWidget {
  const StepCard({
    super.key,
    required this.step,
    required this.index,
    required this.isDone,
    this.onTap,
    this.timerLabel,
  });

  final RecipeStepDto step;
  final int index;
  final bool isDone;
  final VoidCallback? onTap;
  final String? timerLabel;

  @override
  Widget build(BuildContext context) {
    return Card(
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(16),
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Row(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              CircleAvatar(
                backgroundColor: isDone
                    ? Theme.of(context).colorScheme.primary
                    : Colors.grey.shade300,
                foregroundColor: isDone ? Colors.white : Colors.black87,
                child: Text('${index + 1}'),
              ),
              const SizedBox(width: 16),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      step.text,
                      style: TextStyle(
                        decoration:
                            isDone ? TextDecoration.lineThrough : null,
                        color: isDone ? Colors.grey : null,
                      ),
                    ),
                    if (step.timerSeconds > 0) ...[
                      const SizedBox(height: 8),
                      Row(
                        children: [
                          Icon(
                            Icons.timer,
                            size: 16,
                            color: Theme.of(context).colorScheme.secondary,
                          ),
                          const SizedBox(width: 4),
                          Text(
                            timerLabel ?? _formatDuration(step.timerSeconds),
                            style: TextStyle(
                              color: Theme.of(context).colorScheme.secondary,
                              fontWeight: FontWeight.w600,
                            ),
                          ),
                        ],
                      ),
                    ],
                  ],
                ),
              ),
              Icon(
                isDone ? Icons.check_circle : Icons.radio_button_unchecked,
                color: isDone
                    ? Theme.of(context).colorScheme.primary
                    : Colors.grey,
              ),
            ],
          ),
        ),
      ),
    );
  }

  String _formatDuration(int seconds) {
    final m = seconds ~/ 60;
    final s = seconds % 60;
    if (s == 0) return '$m min';
    return '$m:${s.toString().padLeft(2, '0')}';
  }
}