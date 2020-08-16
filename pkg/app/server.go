package app

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/DABronskikh/go-lesson-14/pkg/app/appErr"
	"github.com/DABronskikh/go-lesson-14/pkg/app/dto"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"log"
	"net/http"
	"net/url"
)

type Server struct {
	mux  *http.ServeMux
	ctx  context.Context
	conn *pgxpool.Conn
}

func NewServer(mux *http.ServeMux, ctx context.Context, conn *pgxpool.Conn) *Server {
	return &Server{mux: mux, ctx: ctx, conn: conn}
}

func (s *Server) Init() {
	s.mux.HandleFunc("/getCards", s.getCards)
	s.mux.HandleFunc("/getTransactions", s.getTransactions)
	s.mux.HandleFunc("/getAnalyticSum", s.getAnalyticSum)
	s.mux.HandleFunc("/getAnalyticCategories", s.getAnalyticCategories)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) getCards(w http.ResponseWriter, r *http.Request) {
	userId, ok := url.Parse(r.URL.Query().Get("userId"))
	userIdstr := fmt.Sprintf("%v", userId)
	if ok != nil || userIdstr == "" {
		dtos := dto.ErrDTO{Err: appErr.ErrUserIdNotFound.Error()}
		prepareResponseErr(w, r, dtos)
		return
	}

	cardDB := []*dto.CardDTO{}
	rows, err := s.conn.Query(s.ctx, `
		SELECT id, number, balance, issuer, holder, owner_id, status, created
		FROM cards 
		WHERE owner_id = $1
		LIMIT 50
	`, userIdstr)

	defer rows.Close()

	for rows.Next() {
		cardEl := &dto.CardDTO{}
		err = rows.Scan(&cardEl.Id, &cardEl.Number, &cardEl.Balance, &cardEl.Issuer, &cardEl.Holder, &cardEl.OwnerId, &cardEl.Status, &cardEl.Created)
		if err != nil {
			dtos := dto.ErrDTO{Err: appErr.ErrDB.Error()}
			prepareResponseErr(w, r, dtos)
			return
		}
		cardDB = append(cardDB, cardEl)
	}

	if err != nil {
		if err != pgx.ErrNoRows {
			dtos := dto.ErrDTO{Err: appErr.ErrDB.Error()}
			prepareResponseErr(w, r, dtos)
			return
		}
	}

	prepareResponseCard(w, r, cardDB)
}

func (s *Server) getTransactions(w http.ResponseWriter, r *http.Request) {
	cardId, ok := url.Parse(r.URL.Query().Get("cardId"))
	cardIdstr := fmt.Sprintf("%v", cardId)
	if ok != nil || cardIdstr == "" {
		dtos := dto.ErrDTO{Err: appErr.ErrCardIdNotFound.Error()}
		prepareResponseErr(w, r, dtos)
		return
	}

	transactionDB := []*dto.TransactionDTO{}
	rows, err := s.conn.Query(s.ctx, `
		SELECT id, card_id, amount, created, status, mcc_id, description, supplier_icon_id
		FROM transactions 
		WHERE card_id = $1
		LIMIT 50
	`, cardIdstr)

	defer rows.Close()

	for rows.Next() {
		trEl := &dto.TransactionDTO{}
		err = rows.Scan(&trEl.Id, &trEl.CardId, &trEl.Amount, &trEl.Created, &trEl.Status, &trEl.MccId, &trEl.Description, &trEl.SupplierIconId)
		if err != nil {
			dtos := dto.ErrDTO{Err: appErr.ErrDB.Error()}
			prepareResponseErr(w, r, dtos)
			return
		}
		transactionDB = append(transactionDB, trEl)
	}

	if err != nil {
		if err != pgx.ErrNoRows {
			dtos := dto.ErrDTO{Err: appErr.ErrDB.Error()}
			prepareResponseErr(w, r, dtos)
			return
		}
	}

	prepareResponseTransaction(w, r, transactionDB)
}

func (s *Server) getAnalyticSum(w http.ResponseWriter, r *http.Request) {
	userId, ok := url.Parse(r.URL.Query().Get("userId"))
	userIdstr := fmt.Sprintf("%v", userId)
	if ok != nil || userIdstr == "" {
		dtos := dto.ErrDTO{Err: appErr.ErrUserIdNotFound.Error()}
		prepareResponseErr(w, r, dtos)
		return
	}

	analyticSum := &dto.AnalyticSum{}
	err := s.conn.QueryRow(s.ctx, `
		SELECT m.id, SUM(transactions.amount) AS sum_amount, m.description
		FROM cards
				 LEFT JOIN transactions ON cards.id = transactions.card_id
				 LEFT JOIN mcc m ON transactions.mcc_id = m.id
		WHERE cards.owner_id = $1
		  AND transactions.amount < 0
		GROUP BY m.id, m.description
		ORDER BY sum_amount
		LIMIT 1;
	`, userIdstr).Scan(&analyticSum.MccId, &analyticSum.SumAmount, &analyticSum.Description)
	if err != nil {
		if err != pgx.ErrNoRows {
			log.Println(err)
			dtos := dto.ErrDTO{Err: appErr.ErrDB.Error()}
			prepareResponseErr(w, r, dtos)
			return
		}
	}

	prepareResponseAnalyticSum(w, r, analyticSum)
}

func (s *Server) getAnalyticCategories(w http.ResponseWriter, r *http.Request) {
	userId, ok := url.Parse(r.URL.Query().Get("userId"))
	userIdstr := fmt.Sprintf("%v", userId)
	if ok != nil || userIdstr == "" {
		dtos := dto.ErrDTO{Err: appErr.ErrUserIdNotFound.Error()}
		prepareResponseErr(w, r, dtos)
		return
	}

	analyticCategories := &dto.AnalyticCategories{}
	err := s.conn.QueryRow(s.ctx, `
		SELECT  transactions.mcc_id, count(1) AS col, m.description
		FROM cards
			LEFT JOIN transactions ON cards.id = transactions.card_id
			LEFT JOIN mcc m ON transactions.mcc_id = m.id
		WHERE cards.owner_id = $1
		AND transactions.amount < 0
		GROUP BY transactions.mcc_id, m.description
		ORDER BY col DESC
		LIMIT 1;
	`, userIdstr).Scan(&analyticCategories.MccId, &analyticCategories.Col, &analyticCategories.Description)
	if err != nil {
		if err != pgx.ErrNoRows {
			log.Println(err)
			dtos := dto.ErrDTO{Err: appErr.ErrDB.Error()}
			prepareResponseErr(w, r, dtos)
			return
		}
	}

	prepareResponseAnalyticCategories(w, r, analyticCategories)
}

func prepareResponseCard(w http.ResponseWriter, r *http.Request, dtos []*dto.CardDTO) {
	respBody, err := json.Marshal(dtos)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	_, err = w.Write(respBody)
	if err != nil {
		log.Println(err)
	}
}

func prepareResponseTransaction(w http.ResponseWriter, r *http.Request, dtos []*dto.TransactionDTO) {
	respBody, err := json.Marshal(dtos)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	_, err = w.Write(respBody)
	if err != nil {
		log.Println(err)
	}
}

func prepareResponseAnalyticCategories(w http.ResponseWriter, r *http.Request, dtos *dto.AnalyticCategories) {
	respBody, err := json.Marshal(dtos)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	_, err = w.Write(respBody)
	if err != nil {
		log.Println(err)
	}
}

func prepareResponseAnalyticSum(w http.ResponseWriter, r *http.Request, dtos *dto.AnalyticSum) {
	respBody, err := json.Marshal(dtos)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	_, err = w.Write(respBody)
	if err != nil {
		log.Println(err)
	}
}

func prepareResponseErr(w http.ResponseWriter, r *http.Request, dtos dto.ErrDTO) {
	respBody, err := json.Marshal(dtos)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	_, err = w.Write(respBody)
	if err != nil {
		log.Println(err)
	}
}
