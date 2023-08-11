package repositories

import (
	"database/sql"

	"github.com/nob-swik/go-myapi/models"
)

// 新規投稿をDBにinsertする関数
func InsertArticle(db *sql.DB, article models.Article) (models.Article, error) {
	const sqlStr = `
	insert into articles (title, contents, username, nice, created_at) values
	(?, ?, ?, 0, now());
	`

	result, err := db.Exec(sqlStr, article.Title, article.Contents, article.UserName)
	if err != nil {
		return models.Article{}, err
	}

	var newArticle models.Article
	newArticle.Title, newArticle.Contents, newArticle.UserName = article.Title, article.Contents, article.UserName
	id, _ := result.LastInsertId()
	newArticle.ID = int(id)

	return newArticle, nil
}

// 変数 page で指定されたページに表示する投稿一覧をデータベースから取得する関数
func SelectArticleList(db *sql.DB, page int) ([]models.Article, error) {
	const sqlStr = `
	select article_id, title, contents, username, nice
	from articles
	limit ? offset ?;
	`

	rows, err := db.Query(sqlStr, 5, (page-1)*5)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	articleArray := make([]models.Article, 0)
	for rows.Next() {
		var article models.Article
		rows.Scan(&article.ID, &article.Title, &article.Contents, &article.UserName, &article.NiceNum)

		articleArray = append(articleArray, article)
	}

	return articleArray, nil
}

// 投稿 ID を指定して、記事データを取得する関数
func SelectArticleDetail(db *sql.DB, articleID int) (models.Article, error) {
	const sqlStr = `
	select *
	from articles
	where article_id = ?;
	`

	row := db.QueryRow(sqlStr, articleID)
	if err := row.Err(); err != nil {
		return models.Article{}, err
	}

	var article models.Article
	var createdTime sql.NullTime
	err := row.Scan(&article.ID, &article.Title, &article.Contents, &article.UserName, &article.NiceNum, createdTime)
	if err != nil {
		return models.Article{}, err
	}

	if createdTime.Valid {
		article.CreatedAt = createdTime.Time
	}

	return article, nil
}

// いいねの数を update する関数
func UpdateNiceNum(db *sql.DB, articleID int) error {
	const sqlGetNice = `
	select nice
	from articles
	where article_id = ?;
	`
	const sqlUpdateNice = `update articles set nice = ? where article_id = ?`

	// トランザクションの開始
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	// 現在のいいね数を取得するクエリを実行する
	row := tx.QueryRow(sqlGetNice, articleID)
	if err := row.Err(); err != nil {
		tx.Rollback()
		return err
	}

	// 変数 nicenum に現在のいいね数を読み込む
	var nicenum int
	err = row.Scan(&nicenum)
	if err != nil {
		tx.Rollback()
		return err
	}

	// いいね数を+1 する更新処理を行う
	_, err = tx.Exec(sqlUpdateNice, nicenum+1, articleID)
	if err != nil {
		tx.Rollback()
		return err
	}

	// コミットして処理内容を確定させる
	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}
