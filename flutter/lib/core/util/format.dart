/// Helper di formattazione condivisi.
///
/// Unico posto per formattare durate in secondi come `m:ss` / `m min`,
/// usato sia dal timer di cottura (provider) sia dalle card degli step.
class Format {
  Format._();

  /// Formatta una durata in secondi come `m:ss` (es. `12:05`).
  /// Se i secondi sono 0 restituisce `m min` per compattezza (es. `5 min`).
  static String duration(int totalSeconds) {
    final m = totalSeconds ~/ 60;
    final s = totalSeconds % 60;
    if (s == 0) return '$m min';
    return '$m:${s.toString().padLeft(2, '0')}';
  }
}
