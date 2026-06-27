import 'package:flutter/widgets.dart';

/// Pattern condiviso per le schermate che devono avviare un'azione asincrona
/// al primo frame (tipicamente `ref.read(notifier).load()`) senza farla
/// dentro `build`.
///
/// Sostituisce la ripetizione di
/// `WidgetsBinding.instance.addPostFrameCallback((_) { ... })` presente in
/// più schermate, centralizzando il post-frame load in un solo posto.
void loadOnPostFrame(void Function() action) {
  WidgetsBinding.instance.addPostFrameCallback((_) => action());
}
