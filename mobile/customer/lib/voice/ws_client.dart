import 'dart:async';
import 'dart:convert';

import 'package:web_socket_channel/web_socket_channel.dart';

/// Mobile WS client mirroring DES-0009 §2 frame schema. Mobile **does not**
/// own the switcher state — it just sends/receives the documented control
/// frames and forwards function-call results to whichever Riverpod consumer
/// has subscribed via [events].
///
/// Client→gateway control frames implemented (7 user-functional + `ping`):
/// `start_audio`, `stop_audio`, `mute`, `unmute`, `cancel_speech`,
/// `language_switch`, `quick_action`, `ping`.
class VoiceWsClient {
  VoiceWsClient({
    required this.gatewayUri,
    required this.authToken,
  });

  final Uri gatewayUri;
  final String authToken;

  WebSocketChannel? _channel;
  StreamSubscription? _sub;
  final _events = StreamController<VoiceEvent>.broadcast();

  Stream<VoiceEvent> get events => _events.stream;

  void connect({String? languageHint}) {
    _channel = WebSocketChannel.connect(gatewayUri);
    _sub = _channel!.stream.listen(
      _onFrame,
      onError: (e) => _events.add(VoiceEvent.error('ws_error', '$e')),
      onDone: () => _events.add(const VoiceEvent.closed('client_done')),
    );
    if (languageHint != null) {
      sendLanguageSwitch(languageHint);
    }
  }

  // Client → gateway frames

  void sendStartAudio() => _sendControl('start_audio', const {});
  void sendStopAudio() => _sendControl('stop_audio', const {});
  void sendMute() => _sendControl('mute', const {});
  void sendUnmute() => _sendControl('unmute', const {});
  void sendCancelSpeech() => _sendControl('cancel_speech', const {});
  void sendLanguageSwitch(String bcp47) =>
      _sendControl('language_switch', {'bcp47': bcp47});
  void sendQuickAction(String name) =>
      _sendControl('quick_action', {'name': name});
  void sendPing() => _sendControl('ping', const {});

  void sendAudioFrame(List<int> pcm) {
    _channel?.sink.add(pcm);
  }

  void _sendControl(String frame, Map<String, Object?> payload) {
    final encoded = jsonEncode({'frame': frame, ...payload});
    _channel?.sink.add(encoded);
  }

  void _onFrame(dynamic raw) {
    if (raw is! String) {
      // binary audio out — handed off to a platform audio sink (out of scope).
      _events.add(const VoiceEvent.audioChunk());
      return;
    }
    try {
      final m = jsonDecode(raw) as Map<String, dynamic>;
      final frame = m['frame'] as String? ?? 'unknown';
      _events.add(VoiceEvent.fromFrame(frame, m));
    } catch (e) {
      _events.add(VoiceEvent.error('decode_error', '$e'));
    }
  }

  Future<void> close() async {
    await _sub?.cancel();
    await _channel?.sink.close();
    await _events.close();
  }
}

/// Normalized event surface for screens. One per DES-0009 §2 gateway frame.
class VoiceEvent {
  const VoiceEvent._(this.kind, {this.data = const {}});

  factory VoiceEvent.fromFrame(String frame, Map<String, dynamic> m) {
    switch (frame) {
      case 'session_open':
        return VoiceEvent._(VoiceEventKind.sessionOpen, data: m);
      case 'session_closed':
        return VoiceEvent._(VoiceEventKind.sessionClosed, data: m);
      case 'transcript':
        return VoiceEvent._(VoiceEventKind.transcript, data: m);
      case 'function_call':
        return VoiceEvent._(VoiceEventKind.functionCall, data: m);
      case 'function_result':
        return VoiceEvent._(VoiceEventKind.functionResult, data: m);
      case 'audio_meter':
        return VoiceEvent._(VoiceEventKind.audioMeter, data: m);
      case 'provider_switched':
        return VoiceEvent._(VoiceEventKind.providerSwitched, data: m);
      case 'pong':
        return VoiceEvent._(VoiceEventKind.pong, data: m);
      case 'quota_warning':
        return VoiceEvent._(VoiceEventKind.quotaWarning, data: m);
      case 'error':
      default:
        return VoiceEvent._(VoiceEventKind.error, data: m);
    }
  }

  const VoiceEvent.audioChunk()
      : kind = VoiceEventKind.audioChunk,
        data = const {};
  const VoiceEvent.closed(String reason)
      : kind = VoiceEventKind.sessionClosed,
        data = const {'reason': 'closed'};
  VoiceEvent.error(String code, String message)
      : kind = VoiceEventKind.error,
        data = {'code': code, 'message': message};

  final VoiceEventKind kind;
  final Map<String, dynamic> data;
}

enum VoiceEventKind {
  sessionOpen,
  sessionClosed,
  transcript,
  functionCall,
  functionResult,
  audioMeter,
  providerSwitched,
  pong,
  quotaWarning,
  audioChunk,
  error,
}
