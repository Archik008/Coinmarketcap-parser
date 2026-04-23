package infra

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/callbackquery"

	"crypto_parser/internal/parser/application/ports/in"
	"crypto_parser/internal/parser/domain/dto"
	"crypto_parser/internal/reporting/domain/ports"
	"crypto_parser/internal/reporting/domain/valueobject"
)

const (
	callbackConfirmQuit = "quit:confirm"
	callbackCancelQuit  = "quit:cancel"
)

type Bot struct {
	token         string
	creatorID     int64
	parser        in.ParserUseCase
	reportService ports.ReportGenerator
	cancel        context.CancelFunc
}

func NewBot(token string, creatorID int64, parser in.ParserUseCase, reportService ports.ReportGenerator) *Bot {
	return &Bot{token: token, creatorID: creatorID, parser: parser, reportService: reportService}
}

// Run запускает long-polling и блокируется до отмены ctx.
func (b *Bot) Run(ctx context.Context) error {
	ctx, b.cancel = context.WithCancel(ctx)

	bot, err := gotgbot.NewBot(b.token, nil)
	if err != nil {
		return fmt.Errorf("bot init: %w", err)
	}

	dispatcher := ext.NewDispatcher(&ext.DispatcherOpts{
		UnhandledErrFunc: func(err error) {
			slog.Error("[bot] unhandled error", "err", err)
		},
	})

	dispatcher.AddHandler(handlers.NewCommand("start", b.handleStart))
	dispatcher.AddHandler(handlers.NewCommand("report", b.handleReport))
	dispatcher.AddHandler(handlers.NewCommand("quit", b.handleQuit))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Equal(callbackConfirmQuit), b.handleQuitConfirm))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Equal(callbackCancelQuit), b.handleQuitCancel))

	updater := ext.NewUpdater(dispatcher, nil)
	if err := updater.StartPolling(bot, nil); err != nil {
		return fmt.Errorf("start polling: %w", err)
	}

	slog.Info("[bot] started", "username", bot.User.Username)

	go func() {
		<-ctx.Done()
		updater.Stop()
	}()

	updater.Idle()
	return nil
}

// ── handlers ──────────────────────────────────────────────────────────────────

func (b *Bot) handleStart(bot *gotgbot.Bot, ctx *ext.Context) error {
	_, err := bot.SendMessage(ctx.EffectiveChat.Id,
		"👋 Привет! Я крипто-парсер бот.\n\n"+
			"📊 /report — сгенерировать PDF-отчёт по рынку",
		nil)
	return err
}

func (b *Bot) handleReport(bot *gotgbot.Bot, ectx *ext.Context) error {
	chatID := ectx.EffectiveChat.Id

	notice, _ := bot.SendMessage(chatID, "⏳ Собираю данные и генерирую отчёт...", nil)

	snap, err := b.parser.ParseAll(context.Background())
	if err != nil {
		b.replyError(bot, chatID, "получение данных", err)
		return nil
	}

	data := snapshotToReportData(snap)

	path, err := b.reportService.Generate(data)
	if err != nil {
		b.replyError(bot, chatID, "генерация PDF", err)
		return nil
	}
	defer os.Remove(path)

	f, err := os.Open(path)
	if err != nil {
		b.replyError(bot, chatID, "открытие PDF", err)
		return nil
	}
	defer f.Close()

	caption := fmt.Sprintf(
		"📈 *Crypto Market Report*\n_%s_",
		data.GeneratedAt.Format("02 Jan 2006, 15:04 UTC"),
	)
	_, err = bot.SendDocument(chatID,
		gotgbot.InputFileByReader("report.pdf", f),
		&gotgbot.SendDocumentOpts{Caption: caption, ParseMode: "Markdown"},
	)
	if err != nil {
		return fmt.Errorf("send document: %w", err)
	}

	if notice != nil {
		_, _ = bot.DeleteMessage(chatID, notice.MessageId, nil)
	}

	return nil
}

// handleQuit — шлёт создателю запрос подтверждения через inline-кнопки.
func (b *Bot) handleQuit(bot *gotgbot.Bot, ectx *ext.Context) error {
	if ectx.EffectiveUser.Id != b.creatorID {
		slog.Warn("[bot] unauthorized quit attempt", "user_id", ectx.EffectiveUser.Id)
		return nil
	}

	_, err := bot.SendMessage(ectx.EffectiveChat.Id,
		"⚠️ Ты уверен, что хочешь остановить бота?",
		&gotgbot.SendMessageOpts{
			ReplyMarkup: gotgbot.InlineKeyboardMarkup{
				InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
					{
						{Text: "✅ Да, остановить", CallbackData: callbackConfirmQuit},
						{Text: "❌ Нет", CallbackData: callbackCancelQuit},
					},
				},
			},
		},
	)
	return err
}

// handleQuitConfirm — подтверждение: отвечает на callback и останавливает бота.
func (b *Bot) handleQuitConfirm(bot *gotgbot.Bot, ectx *ext.Context) error {
	cb := ectx.CallbackQuery

	if cb.From.Id != b.creatorID {
		_, _ = cb.Answer(bot, &gotgbot.AnswerCallbackQueryOpts{Text: "Нет доступа."})
		return nil
	}

	_, _ = cb.Answer(bot, &gotgbot.AnswerCallbackQueryOpts{Text: "Останавливаю..."})
	_, _ = bot.DeleteMessage(cb.Message.GetChat().Id, cb.Message.GetMessageId(), nil)

	slog.Info("[bot] shutdown confirmed by creator")
	b.cancel()
	return nil
}

// handleQuitCancel — отмена: удаляет сообщение с кнопками.
func (b *Bot) handleQuitCancel(bot *gotgbot.Bot, ectx *ext.Context) error {
	cb := ectx.CallbackQuery

	_, _ = cb.Answer(bot, &gotgbot.AnswerCallbackQueryOpts{Text: "Отменено."})
	_, _ = bot.DeleteMessage(cb.Message.GetChat().Id, cb.Message.GetMessageId(), nil)

	return nil
}

// ── helpers ───────────────────────────────────────────────────────────────────

func snapshotToReportData(snap in.Snapshot) valueobject.ReportData {
	return valueobject.ReportData{
		GeneratedAt: time.Now().UTC(),
		MarketCap: dto.MarketCapDTO{
			Value:        snap.MarketCap.Value,
			Change24hPct: snap.MarketCap.Change24hPct,
			CapturedAt:   snap.MarketCap.CapturedAt,
		},
		CMC20: dto.CMC20DTO{
			Value:        snap.CMC20.Value,
			Change24hPct: snap.CMC20.Change24hPct,
			CapturedAt:   snap.CMC20.CapturedAt,
		},
		FearGreed: dto.FearGreedDTO{
			Value:      snap.FearGreed.Value,
			Label:      string(snap.FearGreed.Label),
			CapturedAt: snap.FearGreed.CapturedAt,
		},
		AltcoinSeason: dto.AltcoinSeasonDTO{
			Index:        snap.AltcoinSeason.Index,
			Total:        snap.AltcoinSeason.Total,
			Outperformed: snap.AltcoinSeason.Outperformed,
			IsAltSeason:  snap.AltcoinSeason.IsAltSeason,
			CapturedAt:   snap.AltcoinSeason.CapturedAt,
		},
	}
}

func (b *Bot) replyError(bot *gotgbot.Bot, chatID int64, stage string, err error) {
	_, _ = bot.SendMessage(chatID,
		fmt.Sprintf("❌ Ошибка (%s): %v", stage, err),
		nil)
}
