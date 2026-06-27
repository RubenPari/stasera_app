import 'package:flutter_local_notifications/flutter_local_notifications.dart';
import 'package:timezone/data/latest_all.dart' as tz;
import 'package:timezone/timezone.dart' as tz;

/// Servizio per schedulare la notifica giornaliera "Cosa mangi stasera?".
/// Viene schedulata alle 18:30 di ogni giorno feriale (lun-ven).
class NotificationService {
  NotificationService._();

  static final _plugin = FlutterLocalNotificationsPlugin();

  static Future<void> init() async {
    tz.initializeTimeZones();

    const androidInit = AndroidInitializationSettings('@mipmap/ic_launcher');
    const iosInit = DarwinInitializationSettings();
    const settings = InitializationSettings(android: androidInit, iOS: iosInit);

    await _plugin.initialize(settings);
  }

  /// Schedula una notifica ogni lunedì-venerdì alle 18:30.
  /// Sostituisce eventuali pianificazioni precedenti.
  static Future<void> scheduleWeekdayReminders() async {
    await cancelAllReminders();
    for (var weekday = DateTime.monday;
        weekday <= DateTime.friday;
        weekday++) {
      final when = _nextWeekdayAt(hour: 18, minute: 30, weekday: weekday);
      await _plugin.zonedSchedule(
        weekday,
        'Cosa mangi stasera?',
        "Apri l'app e scopri la cena di stasera.",
        when,
        const NotificationDetails(
          android: AndroidNotificationDetails(
            'stasera_weekday',
            'Promemoria serale',
            channelDescription: 'Notifica serale lun-ven alle 18:30',
            importance: Importance.defaultImportance,
          ),
          iOS: DarwinNotificationDetails(),
        ),
        androidScheduleMode: AndroidScheduleMode.inexactAllowWhileIdle,
        uiLocalNotificationDateInterpretation:
            UILocalNotificationDateInterpretation.absoluteTime,
        matchDateTimeComponents: DateTimeComponents.dayOfWeekAndTime,
      );
    }
  }

  /// Cancella tutte le notifiche settimanali pianificate.
  static Future<void> cancelAllReminders() async {
    for (var id = DateTime.monday; id <= DateTime.friday; id++) {
      await _plugin.cancel(id);
    }
  }

  /// Prossima occorrenza futura (oggi se non ancora passata, altrimenti +7gg)
  /// del giorno della settimana [weekday] (1=lunedì ... 7=domenica) all'orario indicato.
  static tz.TZDateTime _nextWeekdayAt({
    required int hour,
    required int minute,
    required int weekday,
  }) {
    final now = tz.TZDateTime.now(tz.local);
    var scheduled = tz.TZDateTime(
      tz.local,
      now.year,
      now.month,
      now.day,
      hour,
      minute,
    );
    while (scheduled.weekday != weekday || !scheduled.isAfter(now)) {
      scheduled = scheduled.add(const Duration(days: 1));
    }
    return scheduled;
  }

  /// Schedula (o rischedula) la notifica giornaliera alle 18:30.
  /// Da chiamare al login e ogni volta che l'utente apre l'app.
  static Future<void> scheduleDailyReminder() async {
    await _plugin.cancel(0);

    final now = tz.TZDateTime.now(tz.local);
    var scheduled = tz.TZDateTime(
      tz.local,
      now.year,
      now.month,
      now.day,
      18,
      30,
    );

    // Se già passato oggi, sposta a domani.
    if (scheduled.isBefore(now)) {
      scheduled = scheduled.add(const Duration(days: 1));
    }

    await _plugin.zonedSchedule(
      0,
      'Cosa mangi stasera?',
      'Apri l\'app e scopri la cena di stasera.',
      scheduled,
      const NotificationDetails(
        android: AndroidNotificationDetails(
          'stasera_daily',
          'Promemoria serale',
          channelDescription: 'Notifica giornaliera alle 18:30',
          importance: Importance.defaultImportance,
        ),
        iOS: DarwinNotificationDetails(),
      ),
      androidScheduleMode: AndroidScheduleMode.inexactAllowWhileIdle,
      uiLocalNotificationDateInterpretation:
          UILocalNotificationDateInterpretation.absoluteTime,
      matchDateTimeComponents: DateTimeComponents.time,
    );
  }

  static Future<void> cancelAll() async {
    await _plugin.cancel(0);
  }
}