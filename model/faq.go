package model

import (
	"fmt"
	"sec-app-server/db"
)

type FAQ struct {
	ID       int    `json:"id"`
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

func GetFAQs() ([]FAQ, error) {
	faqs := []FAQ{}
	sql, err := db.DB.Query("SELECT * FROM faq_question")
	if err != nil {
		fmt.Println("Error fetching FAQs:", err)
		return nil, err
	}
	defer sql.Close()

	for sql.Next() {
		var faq FAQ
		if err := sql.Scan(&faq.ID, &faq.Question, &faq.Answer); err != nil {
			return nil, err
		}
		faqs = append(faqs, faq)
	}

	return faqs, nil
}

func GetFAQ(id string) (*FAQ, error) {
	var faq FAQ
	err := db.DB.QueryRow("SELECT * FROM faq_question WHERE id=$1", id).Scan(&faq);
	if err != nil {
		fmt.Println("Error fetching FAQs:", err)
		return nil, err
	}

	return &faq, nil
}

func AddFAQ(question, answer string) error {
	sql, err := db.DB.Prepare("INSERT INTO faq_question (question, answer) VALUES ($1, $2)")
	if err != nil {
		fmt.Println("Error preparing FAQ insertion:", err)
		return err
	}
	defer sql.Close()

	_, err = sql.Exec(question, answer)
	if err != nil {
		fmt.Println("Error executing FAQ insertion:", err)
		return err
	}

	return nil
}


func DeleteFAQ(id string) error {
	sql, err := db.DB.Prepare("DELETE FROM faq_question WHERE id = $1")
	if err != nil {
		fmt.Println("Error preparing FAQ deletion:", err)
		return err
	}
	defer sql.Close()

	_, err = sql.Exec(id)
	if err != nil {
		fmt.Println("Error executing FAQ deletion:", err)
		return err
	}

	return nil
}

func UpdateFAQ(id, question, answer string) error {
	sql, err := db.DB.Prepare("UPDATE faq_question SET question=$1, answer=$2 WHERE id = $3")
	if err != nil {
		fmt.Println("Error preparing FAQ update:", err)
		return err
	}
	defer sql.Close()

	_, err = sql.Exec(question, answer, id)
	if err != nil {
		fmt.Println("Error executing FAQ update:", err)
		return err
	}

	return nil
}