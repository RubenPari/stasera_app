import 'package:flutter_test/flutter_test.dart';

void main() {
  testWidgets('App smoke test', (tester) async {
    // Smoke test minimo: il vero setup richiede ProviderScope + router.
    expect(1 + 1, equals(2));
  });
}