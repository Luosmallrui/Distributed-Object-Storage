package dao

import (
	"context"
	"fmt"
	"gorm.io/gorm"
	"reflect"
	"strconv"
	"strings"
)

type Base struct {
	DB *gorm.DB
}

func (b *Base) GetDB() *gorm.DB {
	return b.DB
}

// UpdateDB update gorm.DB info
func (b *Base) UpdateDB(db *gorm.DB) {
	b.DB = db
}

func (b *Base) Where(col string, v interface{}) Option {
	return WhereIgZero(col, v)
}

func (b *Base) WhereWithEmpty(col string, v interface{}) Option {
	return Where(col, v)
}

func Where(col string, v interface{}) Option {
	if v == nil {
		return emptyOption
	}
	return optionFunc(func(o *options) { o.query[col] = v })
}

func WhereLike(col string, v string) Option {
	if v == "" {
		return emptyOption
	}
	qv := "%s" + v + "%s"
	return optionFunc(func(o *options) { o.query[fmt.Sprintf("%s like ?", col)] = qv })
}

// WhereIgZero where with ignore zero value
func WhereIgZero(col string, v interface{}) Option {
	if v == nil {
		return emptyOption
	}

	vv := reflect.ValueOf(v)
	vv = reflect.Indirect(vv)

	if vv.IsZero() {
		return emptyOption
	}

	return optionFunc(func(o *options) { o.query[col] = v })
}

func (b *Base) GetByOptionV2(result interface{}, opts ...Option) error {
	options := options{
		query: make(map[string]interface{}, len(opts)),
	}
	for _, o := range opts {
		o.apply(&options)
	}

	tx := b.DB.Model(result)
	if options.ctx != nil {
		tx = tx.WithContext(options.ctx)
	}
	for k, v := range options.query {
		if strings.Contains(k, "?") {
			tx = tx.Where(k, v)
		} else {
			tx = tx.Where(map[string]interface{}{
				k: v,
			})
		}
	}

	return tx.First(result).Error
}

type options struct {
	query map[string]interface{}
	ctx   context.Context
}

// Option overrides behavior of Connect.
type Option interface {
	apply(*options)
}

func WithContext(ctx context.Context) Option {
	return optionFunc(func(o *options) {
		o.ctx = ctx
	})
}

type optionFunc func(*options)

func (f optionFunc) apply(o *options) {
	f(o)
}

var emptyOption = optionFunc(func(o *options) {})

type PageCondition struct {
	currentPage int
	pageSize    int
}

func NewPageCondition(current, pageSize int) *PageCondition {
	return &PageCondition{
		currentPage: current,
		pageSize:    pageSize,
	}
}

func (p *PageCondition) CurrentPage() int {
	if p.currentPage <= 0 {
		p.currentPage = 1
	}
	return p.currentPage
}

func (p *PageCondition) PageSize() int {
	if p.pageSize == 0 {
		p.pageSize = 20
	}
	return p.pageSize
}

func (p PageCondition) Limit() int {
	return p.PageSize()
}

func (p *PageCondition) Offset() int {
	return p.PageSize() * (p.CurrentPage() - 1)
}

// 自定义sql查询
type Condition struct {
	list []*conditionInfo
}

// And a condition by and .and 一个条件
func (c *Condition) And(column string, cases string, value interface{}) {
	c.list = append(c.list, &conditionInfo{
		andor:  "and",
		column: column, // 列名
		case_:  cases,  // 条件(and,or,in,>=,<=)
		value:  value,
	})
}

// Or a condition by or .or 一个条件
func (c *Condition) Or(column string, cases string, value interface{}) {
	c.list = append(c.list, &conditionInfo{
		andor:  "or",
		column: column, // 列名
		case_:  cases,  // 条件(and,or,in,>=,<=)
		value:  value,
	})
}

func (c *Condition) Get() (where string, out []interface{}) {
	firstAnd := -1
	for i := 0; i < len(c.list); i++ { // 查找第一个and
		if c.list[i].andor == "and" {
			where = fmt.Sprintf("`%v` %v ?", c.list[i].column, c.list[i].case_)
			out = append(out, c.list[i].value)
			firstAnd = i
			break
		}
	}

	if firstAnd < 0 && len(c.list) > 0 { // 补刀
		where = fmt.Sprintf("`%v` %v ?", c.list[0].column, c.list[0].case_)
		out = append(out, c.list[0].value)
		firstAnd = 0
	}

	for i := 0; i < len(c.list); i++ { // 添加剩余的
		if firstAnd != i {
			where += fmt.Sprintf(" %v `%v` %v ?", c.list[i].andor, c.list[i].column, c.list[i].case_)
			out = append(out, c.list[i].value)
		}
	}

	return
}

type conditionInfo struct {
	andor  string
	column string // 列名
	case_  string // 条件(in,>=,<=)
	value  interface{}
}

type PagedData struct {
	Results  interface{} `json:"data"`
	Count    int64       `json:"total"`
	Current  int         `form:"current" json:"current"`
	PageSize int         `form:"pageSize" json:"pageSize"`
}

func CountGroups(db *gorm.DB, tableName, groupByColumn string) (map[string]int64, error) {
	rows, err := db.Table(tableName).Group(groupByColumn).Select(groupByColumn + ", count(*) AS count").Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	res := make(map[string]int64)
	for rows.Next() {
		var column, count string
		err = rows.Scan(&column, &count)
		if err != nil {
			return nil, err
		}
		i, err := strconv.ParseInt(count, 10, 64)
		if err != nil {
			// return nil, err
			i = 0
		}
		res[column] = i
	}
	return res, nil
}

func CountByCondition(db *gorm.DB, tableName, where string) (result int64, err error) {
	err = db.Table(tableName).Where(where).Count(&result).Error
	return
}

func CountJoinsByCondition(db *gorm.DB, tableName, join, where string) (result int64, err error) {
	err = db.Table(tableName).Joins(join).Where(where).Count(&result).Error
	return
}

func TopNGroups(db *gorm.DB, tableName, groupByColumn string, limit int) (result []map[string]string, err error) {
	rows, err := db.Table(tableName).Order("total desc").Select("count(*) as total, workspace.name").Joins("INNER JOIN workspace ON workspace_id = workspace.id").Group(groupByColumn).Limit(limit).Rows()
	defer rows.Close()
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	vals := make([][]byte, len(cols))
	scans := make([]interface{}, len(cols))
	for i := range vals {
		scans[i] = &vals[i]
	}
	var results []map[string]string
	for rows.Next() {
		err = rows.Scan(scans...)
		if err != nil {
			return nil, err
		}
		row := make(map[string]string)
		for k, v := range vals {
			key := cols[k]
			row[key] = string(v)
		}
		results = append(results, row)

	}
	return results, nil
}

// GroupByDateWithinOneYear 以天为单位，汇总(指定的数据表)近1年时间的数据
func GroupByDateWithinOneYear(db *gorm.DB, tableName string) (result []map[string]string, err error) {
	rows, err := db.Table(tableName).Select("count(*) as total, DATE(created_at) as date").Where("created_at between date_sub(now(),interval 12 month) and now()").Group("DATE(created_at)").Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	vals := make([][]byte, len(cols))
	scans := make([]interface{}, len(cols))
	for i := range vals {
		scans[i] = &vals[i]
	}
	var results []map[string]string
	for rows.Next() {
		err = rows.Scan(scans...)
		if err != nil {
			return nil, err
		}
		row := make(map[string]string)
		for k, v := range vals {
			key := cols[k]
			row[key] = string(v)
		}
		results = append(results, row)

	}
	return results, nil
}
