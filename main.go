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
	CurrentGame   string
	BetAmount     int
	BlackjackHand []string
	DealerHand    []string
}

var (
	users      = make(map[int64]*User)
	gameStates = make(map[int64]*GameState)
	bot        *tgbotapi.BotAPI
)

// Карточные масти и значения
var suits = []string{"♠", "♥", "♦", "♣"}
var values = []string{"2", "3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K", "A"}

func main() {
	var err error
	bot, err = tgbotapi.NewBotAPI("8075891599:AAE3IUQ3YGpIrEcwsjkBc-rbageCGJ8xX_U")
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
	case "/start":
		sendMessage(userID, "Добро пожаловать в казино-бот! Выберите игру:\n/roulette - Рулетка\n/blackjack - Блэкджек (21)\n/balance - Баланс")
	case "/balance":
		sendBalance(userID)
	case "/roulette":
		startRoulette(userID)
	case "/blackjack":
		startBlackjack(userID)
	default:
		handleGameInput(userID, text)
	}
}

func handleGameInput(userID int64, text string) {
	state, ok := gameStates[userID]
	if !ok {
		return
	}

	switch state.CurrentGame {
	case "roulette":
		handleRouletteBet(userID, text)
	case "blackjack":
		handleBlackjackAction(userID, text)
	}
}

// Рулетка
func startRoulette(userID int64) {
	gameStates[userID] = &GameState{
		CurrentGame: "roulette",
	}

	msg := "Ставки на рулетку!\n\n" +
		"Вы можете ставить на:\n" +
		"- Число (1-36) - выплата 35:1\n" +
		"- Красное/Черное (red/black) - выплата 1:1\n" +
		"- Четное/Нечетное (even/odd) - выплата 1:1\n" +
		"- Первые 12, средние 12, последние 12 (1st12, 2nd12, 3rd12) - выплата 2:1\n\n" +
		"Введите ставку в формате: <сумма> <тип ставки>\n" +
		"Пример: 100 red\n\n" +
		"Ваш баланс: " + strconv.Itoa(users[userID].Balance)

	sendMessage(userID, msg)
}

func handleRouletteBet(userID int64, text string) {
	parts := strings.Split(text, " ")
	if len(parts) != 2 {
		sendMessage(userID, "Неверный формат ставки. Пример: 100 red")
		return
	}

	betAmount, err := strconv.Atoi(parts[0])
	if err != nil || betAmount <= 0 {
		sendMessage(userID, "Неверная сумма ставки")
		return
	}

	user := users[userID]
	if user.Balance < betAmount {
		sendMessage(userID, "Недостаточно средств на балансе")
		return
	}

	betType := strings.ToLower(parts[1])
	validBets := map[string]bool{
		"red": true, "black": true, "even": true, "odd": true,
		"1st12": true, "2nd12": true, "3rd12": true,
	}

	// Проверка числовых ставок (1-36)
	isNumberBet := false
	if num, err := strconv.Atoi(betType); err == nil {
		if num >= 1 && num <= 36 {
			isNumberBet = true
			validBets[betType] = true
		}
	}

	if !validBets[betType] && !isNumberBet {
		sendMessage(userID, "Неверный тип ставки")
		return
	}

	// Спин рулетки
	rand.Seed(time.Now().UnixNano())
	winNumber := rand.Intn(37) // 0-36
	winColor := getColor(winNumber)

	// Определение выигрыша
	won := false
	payout := 0

	switch {
	case isNumberBet && strconv.Itoa(winNumber) == betType:
		won = true
		payout = betAmount * 35
	case betType == "red" && winColor == "red":
		won = true
		payout = betAmount
	case betType == "black" && winColor == "black":
		won = true
		payout = betAmount
	case betType == "even" && winNumber%2 == 0 && winNumber != 0:
		won = true
		payout = betAmount
	case betType == "odd" && winNumber%2 == 1:
		won = true
		payout = betAmount
	case betType == "1st12" && winNumber >= 1 && winNumber <= 12:
		won = true
		payout = betAmount * 2
	case betType == "2nd12" && winNumber >= 13 && winNumber <= 24:
		won = true
		payout = betAmount * 2
	case betType == "3rd12" && winNumber >= 25 && winNumber <= 36:
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
		result = fmt.Sprintf("🎰 Результат: 0 (зеленый)\n")
	}

	if won {
		result += fmt.Sprintf("🎉 Вы выиграли %d! Новый баланс: %d", payout, user.Balance)
	} else {
		result += fmt.Sprintf("😢 Вы проиграли %d. Новый баланс: %d", betAmount, user.Balance)
	}

	sendMessage(userID, result)
	delete(gameStates, userID)
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

// Блэкджек (21)
func startBlackjack(userID int64) {
	user := users[userID]

	// Проверка минимальной ставки
	minBet := 10
	if user.Balance < minBet {
		sendMessage(userID, fmt.Sprintf("Минимальная ставка %d. Недостаточно средств.", minBet))
		return
	}

	gameStates[userID] = &GameState{
		CurrentGame:   "blackjack",
		BlackjackHand: []string{drawCard(), drawCard()},
		DealerHand:    []string{drawCard(), "??"},
	}

	msg := "🎲 Блэкджек (21)\n\n" +
		"Ваши карты: " + strings.Join(gameStates[userID].BlackjackHand, " ") + "\n" +
		"Сумма: " + strconv.Itoa(calculateHand(gameStates[userID].BlackjackHand)) + "\n\n" +
		"Карта дилера: " + gameStates[userID].DealerHand[0] + " ??" + "\n\n" +
		"Выберите действие:\n" +
		"<ставка> - Сделать ставку (мин. " + strconv.Itoa(minBet) + ")\n" +
		"/hit - Взять карту\n" +
		"/stand - Остановиться\n" +
		"/double - Удвоить (если 2 карты)\n\n" +
		"Ваш баланс: " + strconv.Itoa(user.Balance)

	sendMessage(userID, msg)
}

func handleBlackjackAction(userID int64, text string) {
	state := gameStates[userID]
	user := users[userID]

	// Обработка ставки
	if betAmount, err := strconv.Atoi(text); err == nil {
		minBet := 10
		if betAmount < minBet {
			sendMessage(userID, fmt.Sprintf("Минимальная ставка %d", minBet))
			return
		}

		if user.Balance < betAmount {
			sendMessage(userID, "Недостаточно средств")
			return
		}

		state.BetAmount = betAmount
		sendMessage(userID, fmt.Sprintf("Ставка %d принята. Выберите действие: /hit /stand /double", betAmount))
		return
	}

	// Проверка наличия ставки
	if state.BetAmount == 0 {
		sendMessage(userID, "Сначала сделайте ставку")
		return
	}

	switch text {
	case "/hit":
		state.BlackjackHand = append(state.BlackjackHand, drawCard())
		playerTotal := calculateHand(state.BlackjackHand)

		if playerTotal > 21 {
			user.Balance -= state.BetAmount
			sendMessage(userID, fmt.Sprintf("Перебор (%d)! Вы проиграли %d. Новый баланс: %d",
				playerTotal, state.BetAmount, user.Balance))
			delete(gameStates, userID)
			return
		}

		msg := "Ваши карты: " + strings.Join(state.BlackjackHand, " ") + "\n" +
			"Сумма: " + strconv.Itoa(playerTotal) + "\n\n" +
			"Выберите действие:\n/hit /stand"

		if len(state.BlackjackHand) == 2 {
			msg += " /double"
		}

		sendMessage(userID, msg)

	case "/stand":
		completeBlackjackGame(userID)

	case "/double":
		if len(state.BlackjackHand) != 2 {
			sendMessage(userID, "Удвоение возможно только при 2 картах")
			return
		}

		if user.Balance < state.BetAmount*2 {
			sendMessage(userID, "Недостаточно средств для удвоения")
			return
		}

		state.BetAmount *= 2
		state.BlackjackHand = append(state.BlackjackHand, drawCard())
		playerTotal := calculateHand(state.BlackjackHand)

		if playerTotal > 21 {
			user.Balance -= state.BetAmount
			sendMessage(userID, fmt.Sprintf("Перебор (%d) после удвоения! Вы проиграли %d. Новый баланс: %d",
				playerTotal, state.BetAmount, user.Balance))
		} else {
			completeBlackjackGame(userID)
		}

		delete(gameStates, userID)
	}
}

func completeBlackjackGame(userID int64) {
	state := gameStates[userID]
	user := users[userID]

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
		result = fmt.Sprintf("Перебор (%d)! Вы проиграли %d.", playerTotal, state.BetAmount)
		user.Balance -= state.BetAmount
	} else if dealerTotal > 21 {
		result = fmt.Sprintf("Дилер перебрал (%d)! Вы выиграли %d.", dealerTotal, state.BetAmount)
		user.Balance += state.BetAmount
	} else if playerTotal > dealerTotal {
		result = fmt.Sprintf("Вы победили (%d против %d)! Выигрыш %d.", playerTotal, dealerTotal, state.BetAmount)
		user.Balance += state.BetAmount
	} else if playerTotal == dealerTotal {
		result = fmt.Sprintf("Ничья (%d против %d). Ставка возвращена.", playerTotal, dealerTotal)
	} else {
		result = fmt.Sprintf("Вы проиграли (%d против %d). Потеря %d.", playerTotal, dealerTotal, state.BetAmount)
		user.Balance -= state.BetAmount
	}

	// Формирование сообщения с результатом
	msg := "🎲 Результат игры:\n\n" +
		"Ваши карты: " + strings.Join(state.BlackjackHand, " ") + " = " + strconv.Itoa(playerTotal) + "\n" +
		"Карты дилера: " + strings.Join(state.DealerHand, " ") + " = " + strconv.Itoa(dealerTotal) + "\n\n" +
		result + "\n" +
		"Новый баланс: " + strconv.Itoa(user.Balance)

	sendMessage(userID, msg)
	delete(gameStates, userID)
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

// Вспомогательные функции
func sendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	bot.Send(msg)
}

func sendBalance(userID int64) {
	user := users[userID]
	sendMessage(userID, fmt.Sprintf("Ваш баланс: %d", user.Balance))
}
