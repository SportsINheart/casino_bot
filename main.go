package main

import (
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// User представляет информацию о пользователе
type User struct {
	ID       int64
	Username string
	Balance  int
}

// GameState представляет состояние игры
type GameState struct {
	CurrentGame     string
	BetAmount       int
	BlackjackHand   []string
	DealerHand      []string
	RouletteBetType string
	DiceBetType     string
}

var (
	users      = make(map[int64]*User)
	gameStates = make(map[int64]*GameState)
	bot        *tgbotapi.BotAPI

	// Основная клавиатура
	mainKeyboard = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🎰 Рулетка"),
			tgbotapi.NewKeyboardButton("🎲 Блэкджек"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🎲 Кости"),
			tgbotapi.NewKeyboardButton("🎰 Слоты"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("💰 Баланс"),
			tgbotapi.NewKeyboardButton("❓ Помощь"),
		),
	)

	// Клавиатура для рулетки
	rouletteKeyboard = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🔴 Красное"),
			tgbotapi.NewKeyboardButton("⚫ Черное"),
			tgbotapi.NewKeyboardButton("🟢 Зеро"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("⚪ Четное"),
			tgbotapi.NewKeyboardButton("⚫ Нечетное"),
			tgbotapi.NewKeyboardButton("1-12"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("13-24"),
			tgbotapi.NewKeyboardButton("25-36"),
			tgbotapi.NewKeyboardButton("🔙 Назад"),
		),
	)

	// Клавиатура для блэкджека
	blackjackKeyboard = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("⬇️ Взять"),
			tgbotapi.NewKeyboardButton("✋ Стоять"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("💰 Удвоить"),
			tgbotapi.NewKeyboardButton("🔙 Назад"),
		),
	)

	// Клавиатура для костей
	diceKeyboard = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🎲 Четное"),
			tgbotapi.NewKeyboardButton("🎲 Нечетное"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🎲 <7"),
			tgbotapi.NewKeyboardButton("🎲 =7"),
			tgbotapi.NewKeyboardButton("🎲 >7"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🔙 Назад"),
		),
	)

	// Клавиатура для слотов
	slotsKeyboard = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🎰 Крутить"),
			tgbotapi.NewKeyboardButton("🔙 Назад"),
		),
	)

	// Клавиатура для ставок
	betKeyboard = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("10"),
			tgbotapi.NewKeyboardButton("50"),
			tgbotapi.NewKeyboardButton("100"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("200"),
			tgbotapi.NewKeyboardButton("500"),
			tgbotapi.NewKeyboardButton("🔙 Назад"),
		),
	)

	// Карточные масти и значения
	suits  = []string{"♠", "♥", "♦", "♣"}
	values = []string{"2", "3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K", "A"}

	// Символы для слотов
	slotSymbols = []string{"🍒", "🍋", "🍊", "🍇", "🍉", "7️⃣", "🔔", "💎"}
)

func main() {
	var err error
	bot, err = tgbotapi.NewBotAPI("7907157167:AAFbanlT69HoZ_67xyKf3scxD_A_gf9nRjI")
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		handleMessage(update.Message)
	}
}

func handleMessage(msg *tgbotapi.Message) {
	userID := msg.Chat.ID
	text := msg.Text

	// Инициализация пользователя при первом сообщении
	if _, ok := users[userID]; !ok {
		users[userID] = &User{
			ID:       userID,
			Username: msg.From.UserName,
			Balance:  1000, // Начальный баланс
		}
	}

	// Обработка команд
	switch text {
	case "/start", "❓ Помощь", "🔙 Назад":
		sendMainMenu(userID)
	case "💰 Баланс":
		sendBalance(userID)
	case "🎰 Рулетка":
		startRoulette(userID)
	case "🎲 Блэкджек":
		startBlackjack(userID)
	case "🎲 Кости":
		startDice(userID)
	case "🎰 Слоты":
		startSlots(userID)
	default:
		handleGameInput(userID, text)
	}
}

func sendMainMenu(chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "🎰 Добро пожаловать в Casino Bot! Выберите игру:")
	msg.ReplyMarkup = mainKeyboard
	bot.Send(msg)
}

func sendBalance(chatID int64) {
	user := users[chatID]
	msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("💰 Ваш баланс: %d", user.Balance))
	msg.ReplyMarkup = mainKeyboard
	bot.Send(msg)
}

// Рулетка
func startRoulette(chatID int64) {
	gameStates[chatID] = &GameState{
		CurrentGame: "roulette",
	}

	msg := tgbotapi.NewMessage(chatID, "🎰 Выберите тип ставки в рулетке:")
	msg.ReplyMarkup = rouletteKeyboard
	bot.Send(msg)
}

func handleRouletteBet(chatID int64, betType string) {
	state := gameStates[chatID]
	state.RouletteBetType = betType

	msg := tgbotapi.NewMessage(chatID, "💰 Введите сумму ставки или выберите из предложенных:")
	msg.ReplyMarkup = betKeyboard
	bot.Send(msg)
}

func processRouletteBet(chatID int64, betAmount int) {
	state := gameStates[chatID]
	user := users[chatID]

	if user.Balance < betAmount {
		sendMessage(chatID, "⚠️ Недостаточно средств на балансе")
		return
	}

	// Спин рулетки
	rand.Seed(time.Now().UnixNano())
	winNumber := rand.Intn(37) // 0-36
	winColor := getColor(winNumber)

	// Определение выигрыша
	won := false
	payout := 0
	betType := state.RouletteBetType

	switch {
	case betType == "🟢 Зеро" && winNumber == 0:
		won = true
		payout = betAmount * 35
	case betType == "🔴 Красное" && winColor == "red":
		won = true
		payout = betAmount
	case betType == "⚫ Черное" && winColor == "black":
		won = true
		payout = betAmount
	case betType == "⚪ Четное" && winNumber%2 == 0 && winNumber != 0:
		won = true
		payout = betAmount
	case betType == "⚫ Нечетное" && winNumber%2 == 1:
		won = true
		payout = betAmount
	case betType == "1-12" && winNumber >= 1 && winNumber <= 12:
		won = true
		payout = betAmount * 2
	case betType == "13-24" && winNumber >= 13 && winNumber <= 24:
		won = true
		payout = betAmount * 2
	case betType == "25-36" && winNumber >= 25 && winNumber <= 36:
		won = true
		payout = betAmount * 2
	}

	// Обновление баланса
	if won {
		user.Balance += payout
	} else {
		user.Balance -= betAmount
	}

	// Формирование результата
	result := fmt.Sprintf("🎰 Результат: %d %s\n", winNumber, winColor)
	if winNumber == 0 {
		result = "🎰 Результат: 0 (зеленый)\n"
	}

	if won {
		result += fmt.Sprintf("🎉 Вы выиграли %d! Новый баланс: %d", payout, user.Balance)
	} else {
		result += fmt.Sprintf("😢 Вы проиграли %d. Новый баланс: %d", betAmount, user.Balance)
	}

	sendMessageWithKeyboard(chatID, result, mainKeyboard)
	delete(gameStates, chatID)
}

// Блэкджек
func startBlackjack(chatID int64) {
	user := users[chatID]
	minBet := 10

	if user.Balance < minBet {
		sendMessage(chatID, fmt.Sprintf("⚠️ Минимальная ставка %d. Недостаточно средств.", minBet))
		return
	}

	gameStates[chatID] = &GameState{
		CurrentGame:   "blackjack",
		BlackjackHand: []string{drawCard(), drawCard()},
		DealerHand:    []string{drawCard(), "??"},
	}

	msg := fmt.Sprintf("🎲 Блэкджек\n\nВаши карты: %s\nСумма: %d\n\nКарта дилера: %s ??\n\nСделайте ставку:",
		strings.Join(gameStates[chatID].BlackjackHand, " "),
		calculateHand(gameStates[chatID].BlackjackHand),
		gameStates[chatID].DealerHand[0])

	sendMessageWithKeyboard(chatID, msg, betKeyboard)
}

func handleBlackjackAction(chatID int64, text string) {
	state := gameStates[chatID]
	user := users[chatID]

	// Обработка ставки
	if betAmount, err := strconv.Atoi(text); err == nil {
		minBet := 10
		if betAmount < minBet {
			sendMessage(chatID, fmt.Sprintf("⚠️ Минимальная ставка %d", minBet))
			return
		}

		if user.Balance < betAmount {
			sendMessage(chatID, "⚠️ Недостаточно средств")
			return
		}

		state.BetAmount = betAmount
		showBlackjackOptions(chatID)
		return
	}

	// Проверка наличия ставки
	if state.BetAmount == 0 {
		sendMessage(chatID, "⚠️ Сначала сделайте ставку")
		return
	}

	switch text {
	case "⬇️ Взять":
		state.BlackjackHand = append(state.BlackjackHand, drawCard())
		playerTotal := calculateHand(state.BlackjackHand)

		if playerTotal > 21 {
			user.Balance -= state.BetAmount
			sendMessageWithKeyboard(chatID,
				fmt.Sprintf("💥 Перебор (%d)! Вы проиграли %d. Новый баланс: %d",
					playerTotal, state.BetAmount, user.Balance),
				mainKeyboard)
			delete(gameStates, chatID)
			return
		}

		showBlackjackOptions(chatID)

	case "✋ Стоять":
		completeBlackjackGame(chatID)

	case "💰 Удвоить":
		if len(state.BlackjackHand) != 2 {
			sendMessage(chatID, "⚠️ Удвоение возможно только при 2 картах")
			return
		}

		if user.Balance < state.BetAmount*2 {
			sendMessage(chatID, "⚠️ Недостаточно средств для удвоения")
			return
		}

		state.BetAmount *= 2
		state.BlackjackHand = append(state.BlackjackHand, drawCard())
		playerTotal := calculateHand(state.BlackjackHand)

		if playerTotal > 21 {
			user.Balance -= state.BetAmount
			sendMessageWithKeyboard(chatID,
				fmt.Sprintf("💥 Перебор (%d) после удвоения! Вы проиграли %d. Новый баланс: %d",
					playerTotal, state.BetAmount, user.Balance),
				mainKeyboard)
		} else {
			completeBlackjackGame(chatID)
		}

		delete(gameStates, chatID)
	}
}

// Кости
func startDice(chatID int64) {
	gameStates[chatID] = &GameState{
		CurrentGame: "dice",
	}

	msg := tgbotapi.NewMessage(chatID, "🎲 Выберите тип ставки в костях:")
	msg.ReplyMarkup = diceKeyboard
	bot.Send(msg)
}

func handleDiceBet(chatID int64, betType string) {
	state := gameStates[chatID]
	state.DiceBetType = betType

	msg := tgbotapi.NewMessage(chatID, "💰 Введите сумму ставки или выберите из предложенных:")
	msg.ReplyMarkup = betKeyboard
	bot.Send(msg)
}

func processDiceBet(chatID int64, betAmount int) {
	state := gameStates[chatID]
	user := users[chatID]

	if user.Balance < betAmount {
		sendMessage(chatID, "⚠️ Недостаточно средств на балансе")
		return
	}

	// Бросок костей
	rand.Seed(time.Now().UnixNano())
	dice1 := rand.Intn(6) + 1
	dice2 := rand.Intn(6) + 1
	total := dice1 + dice2

	// Определение выигрыша
	won := false
	payout := 0
	betType := state.DiceBetType

	switch {
	case betType == "🎲 Четное" && total%2 == 0:
		won = true
		payout = betAmount
	case betType == "🎲 Нечетное" && total%2 == 1:
		won = true
		payout = betAmount
	case betType == "🎲 <7" && total < 7:
		won = true
		payout = betAmount
	case betType == "🎲 =7" && total == 7:
		won = true
		payout = betAmount * 4
	case betType == "🎲 >7" && total > 7:
		won = true
		payout = betAmount
	}

	// Обновление баланса
	if won {
		user.Balance += payout
	} else {
		user.Balance -= betAmount
	}

	// Формирование результата
	result := fmt.Sprintf("🎲 Результат: %d и %d (сумма: %d)\n", dice1, dice2, total)

	if won {
		result += fmt.Sprintf("🎉 Вы выиграли %d! Новый баланс: %d", payout, user.Balance)
	} else {
		result += fmt.Sprintf("😢 Вы проиграли %d. Новый баланс: %d", betAmount, user.Balance)
	}

	sendMessageWithKeyboard(chatID, result, mainKeyboard)
	delete(gameStates, chatID)
}

// Слоты
func startSlots(chatID int64) {
	gameStates[chatID] = &GameState{
		CurrentGame: "slots",
	}

	msg := tgbotapi.NewMessage(chatID, "🎰 Добро пожаловать в слоты!\nМинимальная ставка: 10\n\nВыберите сумму ставки:")
	msg.ReplyMarkup = betKeyboard
	bot.Send(msg)
}

func processSlotsBet(chatID int64, betAmount int) {
	state := gameStates[chatID]
	user := users[chatID]

	if user.Balance < betAmount {
		sendMessage(chatID, "⚠️ Недостаточно средств на балансе")
		return
	}

	if betAmount < 10 {
		sendMessage(chatID, "⚠️ Минимальная ставка 10")
		return
	}

	state.BetAmount = betAmount

	// Крутим слоты
	rand.Seed(time.Now().UnixNano())
	reels := make([]string, 3)
	for i := 0; i < 3; i++ {
		reels[i] = slotSymbols[rand.Intn(len(slotSymbols))]
	}

	// Проверка выигрыша
	won := false
	payout := 0

	// Все три одинаковые
	if reels[0] == reels[1] && reels[1] == reels[2] {
		won = true
		switch reels[0] {
		case "7️⃣":
			payout = betAmount * 100 // Джекпот за три семерки
		case "💎":
			payout = betAmount * 50
		case "🔔":
			payout = betAmount * 20
		default:
			payout = betAmount * 10
		}
	} else if reels[0] == reels[1] || reels[1] == reels[2] || reels[0] == reels[2] {
		// Две одинаковые
		won = true
		payout = betAmount
	}

	// Обновление баланса
	if won {
		user.Balance += payout
	} else {
		user.Balance -= betAmount
	}

	// Формирование результата
	result := fmt.Sprintf("🎰 [ %s | %s | %s ]\n", reels[0], reels[1], reels[2])

	if won {
		result += fmt.Sprintf("🎉 Вы выиграли %d! Новый баланс: %d", payout, user.Balance)
	} else {
		result += fmt.Sprintf("😢 Вы проиграли %d. Новый баланс: %d", betAmount, user.Balance)
	}

	msg := tgbotapi.NewMessage(chatID, result)
	msg.ReplyMarkup = slotsKeyboard
	bot.Send(msg)
}

func showBlackjackOptions(chatID int64) {
	state := gameStates[chatID]
	playerTotal := calculateHand(state.BlackjackHand)

	msg := fmt.Sprintf("🎲 Ваши карты: %s\nСумма: %d\n\nВыберите действие:",
		strings.Join(state.BlackjackHand, " "),
		playerTotal)

	sendMessageWithKeyboard(chatID, msg, blackjackKeyboard)
}

func completeBlackjackGame(chatID int64) {
	state := gameStates[chatID]
	user := users[chatID]

	// Открываем карту дилера
	state.DealerHand[1] = drawCard()
	dealerTotal := calculateHand(state.DealerHand)

	// Дилер добирает карты до 17
	for dealerTotal < 17 {
		state.DealerHand = append(state.DealerHand, drawCard())
		dealerTotal = calculateHand(state.DealerHand)
	}

	playerTotal := calculateHand(state.BlackjackHand)

	// Определение результата
	result := ""
	if playerTotal > 21 {
		result = fmt.Sprintf("💥 Перебор (%d)! Вы проиграли %d.", playerTotal, state.BetAmount)
		user.Balance -= state.BetAmount
	} else if dealerTotal > 21 {
		result = fmt.Sprintf("🎉 Дилер перебрал (%d)! Вы выиграли %d.", dealerTotal, state.BetAmount)
		user.Balance += state.BetAmount
	} else if playerTotal > dealerTotal {
		result = fmt.Sprintf("🎉 Вы победили (%d против %d)! Выигрыш %d.", playerTotal, dealerTotal, state.BetAmount)
		user.Balance += state.BetAmount
	} else if playerTotal == dealerTotal {
		result = fmt.Sprintf("🤝 Ничья (%d против %d). Ставка возвращена.", playerTotal, dealerTotal)
	} else {
		result = fmt.Sprintf("😢 Вы проиграли (%d против %d). Потеря %d.", playerTotal, dealerTotal, state.BetAmount)
		user.Balance -= state.BetAmount
	}

	// Формирование сообщения с результатом
	msg := "🎲 Результат игры:\n\n" +
		"Ваши карты: " + strings.Join(state.BlackjackHand, " ") + " = " + strconv.Itoa(playerTotal) + "\n" +
		"Карты дилера: " + strings.Join(state.DealerHand, " ") + " = " + strconv.Itoa(dealerTotal) + "\n\n" +
		result + "\n" +
		"Новый баланс: " + strconv.Itoa(user.Balance)

	sendMessageWithKeyboard(chatID, msg, mainKeyboard)
	delete(gameStates, chatID)
}

func drawCard() string {
	rand.Seed(time.Now().UnixNano())
	value := values[rand.Intn(len(values))]
	suit := suits[rand.Intn(len(suits))]
	return value + suit
}

func calculateHand(hand []string) int {
	total := 0
	aces := 0

	for _, card := range hand {
		if card == "??" {
			continue // карта дилера скрыта
		}

		value := strings.TrimRight(card, "♠♥♦♣")
		switch value {
		case "A":
			total += 11
			aces++
		case "K", "Q", "J":
			total += 10
		default:
			if num, err := strconv.Atoi(value); err == nil {
				total += num
			}
		}
	}

	// Учитываем тузы как 1, если перебор
	for total > 21 && aces > 0 {
		total -= 10
		aces--
	}

	return total
}

func getColor(number int) string {
	if number == 0 {
		return "green"
	}

	redNumbers := []int{1, 3, 5, 7, 9, 12, 14, 16, 18, 19, 21, 23, 25, 27, 30, 32, 34, 36}
	for _, n := range redNumbers {
		if n == number {
			return "red"
		}
	}
	return "black"
}

func handleGameInput(chatID int64, text string) {
	state, ok := gameStates[chatID]
	if !ok {
		return
	}

	switch state.CurrentGame {
	case "roulette":
		switch text {
		case "🔴 Красное", "⚫ Черное", "🟢 Зеро", "⚪ Четное", "⚫ Нечетное", "1-12", "13-24", "25-36":
			handleRouletteBet(chatID, text)
		case "10", "50", "100", "200", "500":
			if betAmount, err := strconv.Atoi(text); err == nil {
				processRouletteBet(chatID, betAmount)
			}
		default:
			if betAmount, err := strconv.Atoi(text); err == nil {
				processRouletteBet(chatID, betAmount)
			} else {
				sendMessage(chatID, "⚠️ Неверная сумма ставки")
			}
		}
	case "blackjack":
		handleBlackjackAction(chatID, text)
	case "dice":
		switch text {
		case "🎲 Четное", "🎲 Нечетное", "🎲 <7", "🎲 =7", "🎲 >7":
			handleDiceBet(chatID, text)
		case "10", "50", "100", "200", "500":
			if betAmount, err := strconv.Atoi(text); err == nil {
				processDiceBet(chatID, betAmount)
			}
		default:
			if betAmount, err := strconv.Atoi(text); err == nil {
				processDiceBet(chatID, betAmount)
			} else {
				sendMessage(chatID, "⚠️ Неверная сумма ставки")
			}
		}
	case "slots":
		switch text {
		case "🎰 Крутить":
			if state.BetAmount > 0 {
				processSlotsBet(chatID, state.BetAmount)
			} else {
				sendMessage(chatID, "💰 Введите сумму ставки или выберите из предложенных:")
			}
		case "10", "50", "100", "200", "500":
			if betAmount, err := strconv.Atoi(text); err == nil {
				processSlotsBet(chatID, betAmount)
			}
		default:
			if betAmount, err := strconv.Atoi(text); err == nil {
				processSlotsBet(chatID, betAmount)
			} else {
				sendMessage(chatID, "⚠️ Неверная сумма ставки")
			}
		}
	}
}

// Вспомогательные функции
func sendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	bot.Send(msg)
}

func sendMessageWithKeyboard(chatID int64, text string, keyboard tgbotapi.ReplyKeyboardMarkup) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard
	bot.Send(msg)
}
