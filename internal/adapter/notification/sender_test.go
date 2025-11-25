package notification

import (
	"errors"
	"github.com/beliaev-aa/notifications/internal/domain/port"
	"github.com/beliaev-aa/notifications/tests/mocks"
	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"
	"testing"
)

func TestNewSender(t *testing.T) {
	type testCase struct {
		name           string
		logger         *logrus.Logger
		checkInterface bool
		checkNil       bool
	}

	testCases := []testCase{
		{
			name:           "Create_Sender_Success",
			logger:         logrus.New(),
			checkInterface: true,
			checkNil:       false,
		},
		{
			name:           "Create_Sender_With_Nil_Logger",
			logger:         nil,
			checkInterface: true,
			checkNil:       false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sender := NewSender(tc.logger)

			if tc.checkNil {
				if sender != nil {
					t.Error("expected sender to be nil, got: not nil")
				}
			} else {
				if sender == nil {
					t.Error("expected sender to be created, got: nil")
					return
				}

				if tc.checkInterface {
					if _, ok := sender.(port.NotificationSender); !ok {
						t.Error("expected sender to implement NotificationSender interface")
					}
				}
			}
		})
	}
}

func TestSender_Send(t *testing.T) {
	type testCase struct {
		name             string
		channel          string
		message          string
		registerChannel  bool
		channelName      string
		channelError     error
		expectedError    bool
		expectedErrorMsg string
	}

	testCases := []testCase{
		{
			name:            "Send_Success",
			channel:         "test_channel",
			message:         "Test message",
			registerChannel: true,
			channelName:     "test_channel",
			channelError:    nil,
			expectedError:   false,
		},
		{
			name:             "Send_Empty_Message",
			channel:          "test_channel",
			message:          "",
			registerChannel:  true,
			channelName:      "test_channel",
			channelError:     nil,
			expectedError:    true,
			expectedErrorMsg: "formatted message cannot be empty",
		},
		{
			name:             "Send_Unregistered_Channel",
			channel:          "unregistered_channel",
			message:          "Test message",
			registerChannel:  false,
			channelName:      "",
			channelError:     nil,
			expectedError:    true,
			expectedErrorMsg: "channel 'unregistered_channel' is not registered",
		},
		{
			name:             "Send_Channel_Error",
			channel:          "test_channel",
			message:          "Test message",
			registerChannel:  true,
			channelName:      "test_channel",
			channelError:     errors.New("channel send error"),
			expectedError:    true,
			expectedErrorMsg: "channel send error",
		},
		{
			name:            "Send_With_Whitespace_Message",
			channel:         "test_channel",
			message:         "   ",
			registerChannel: true,
			channelName:     "test_channel",
			channelError:    nil,
			expectedError:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel)
			sender := NewSender(logger).(*Sender)

			if tc.registerChannel {
				mockChannel := mocks.NewMockNotificationChannel(ctrl)
				mockChannel.EXPECT().Channel().Return(tc.channelName).AnyTimes()
				if tc.message != "" || tc.name == "Send_With_Whitespace_Message" {
					sendCall := mockChannel.EXPECT().Send(tc.message)
					if tc.channelError != nil {
						sendCall.Return(tc.channelError)
					} else {
						sendCall.Return(nil)
					}
				}
				sender.RegisterChannel(mockChannel)
			}

			err := sender.Send(tc.channel, tc.message)

			if tc.expectedError {
				if err == nil {
					t.Error("expected error, got: nil")
				} else if tc.expectedErrorMsg != "" && err.Error() != tc.expectedErrorMsg {
					t.Errorf("expected error message %q, got: %q", tc.expectedErrorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestSender_RegisterChannel(t *testing.T) {
	type testCase struct {
		name            string
		channelName     string
		channelNil      bool
		channelError    error
		expectedLog     bool
		checkRegistered bool
	}

	testCases := []testCase{
		{
			name:            "RegisterChannel_Success",
			channelName:     "test_channel",
			channelNil:      false,
			channelError:    nil,
			expectedLog:     true,
			checkRegistered: true,
		},
		{
			name:            "RegisterChannel_With_Nil_Channel",
			channelName:     "",
			channelNil:      true,
			channelError:    nil,
			expectedLog:     false,
			checkRegistered: false,
		},
		{
			name:            "RegisterChannel_With_Empty_Name",
			channelName:     "",
			channelNil:      false,
			channelError:    nil,
			expectedLog:     false,
			checkRegistered: false,
		},
		{
			name:            "RegisterChannel_Multiple_Channels",
			channelName:     "channel1",
			channelNil:      false,
			channelError:    nil,
			expectedLog:     true,
			checkRegistered: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel)
			sender := NewSender(logger).(*Sender)

			if tc.name == "RegisterChannel_Multiple_Channels" {
				mockChannel1 := mocks.NewMockNotificationChannel(ctrl)
				mockChannel1.EXPECT().Channel().Return("channel1").AnyTimes()
				sender.RegisterChannel(mockChannel1)

				mockChannel2 := mocks.NewMockNotificationChannel(ctrl)
				mockChannel2.EXPECT().Channel().Return("channel2").AnyTimes()
				sender.RegisterChannel(mockChannel2)

				if len(sender.channels) != 2 {
					t.Errorf("expected 2 channels registered, got: %d", len(sender.channels))
				}
				return
			}

			var mockChannel *mocks.MockNotificationChannel
			if !tc.channelNil {
				mockChannel = mocks.NewMockNotificationChannel(ctrl)
				mockChannel.EXPECT().Channel().Return(tc.channelName).AnyTimes()
			}

			if tc.channelNil {
				sender.RegisterChannel(nil)
			} else {
				sender.RegisterChannel(mockChannel)
			}

			if tc.checkRegistered {
				if len(sender.channels) != 1 {
					t.Errorf("expected 1 channel registered, got: %d", len(sender.channels))
				}
				if _, exists := sender.channels[tc.channelName]; !exists {
					t.Errorf("expected channel %q to be registered", tc.channelName)
				}
			} else {
				if len(sender.channels) != 0 {
					t.Errorf("expected 0 channels registered, got: %d", len(sender.channels))
				}
			}
		})
	}
}

func TestSender_Integration(t *testing.T) {
	type testCase struct {
		name      string
		channels  []string
		messages  []string
		checkSend bool
	}

	testCases := []testCase{
		{
			name:      "Full_Integration_Test",
			channels:  []string{"channel1", "channel2"},
			messages:  []string{"Message 1", "Message 2"},
			checkSend: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel)
			sender := NewSender(logger).(*Sender)

			for i, channelName := range tc.channels {
				mockChannel := mocks.NewMockNotificationChannel(ctrl)
				mockChannel.EXPECT().Channel().Return(channelName).AnyTimes()
				if tc.checkSend {
					mockChannel.EXPECT().Send(tc.messages[i]).Return(nil)
				}
				sender.RegisterChannel(mockChannel)
			}

			if len(sender.channels) != len(tc.channels) {
				t.Errorf("expected %d channels registered, got: %d", len(tc.channels), len(sender.channels))
			}

			if tc.checkSend {
				for i, channelName := range tc.channels {
					err := sender.Send(channelName, tc.messages[i])
					if err != nil {
						t.Errorf("unexpected error sending to channel %q: %v", channelName, err)
					}
				}
			}
		})
	}
}

func TestSender_EdgeCases(t *testing.T) {
	type testCase struct {
		name          string
		channel       string
		message       string
		expectedError bool
	}

	testCases := []testCase{
		{
			name:          "Send_With_Long_Message",
			channel:       "test_channel",
			message:       "This is a very long message " + string(make([]byte, 1000)),
			expectedError: false,
		},
		{
			name:          "Send_With_Special_Characters",
			channel:       "test_channel",
			message:       "Message with special chars: !@#$%^&*()",
			expectedError: false,
		},
		{
			name:          "Send_With_Unicode",
			channel:       "test_channel",
			message:       "Сообщение с кириллицей 🚀",
			expectedError: false,
		},
		{
			name:          "Send_With_Newlines",
			channel:       "test_channel",
			message:       "Line 1\nLine 2\nLine 3",
			expectedError: false,
		},
		{
			name:          "Send_With_JSON_Message",
			channel:       "test_channel",
			message:       `{"key": "value", "number": 123}`,
			expectedError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel)
			sender := NewSender(logger).(*Sender)

			mockChannel := mocks.NewMockNotificationChannel(ctrl)
			mockChannel.EXPECT().Channel().Return(tc.channel).AnyTimes()
			mockChannel.EXPECT().Send(tc.message).Return(nil)
			sender.RegisterChannel(mockChannel)

			err := sender.Send(tc.channel, tc.message)

			if tc.expectedError {
				if err == nil {
					t.Error("expected error, got: nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}
