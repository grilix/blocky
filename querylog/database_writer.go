package querylog

import (
	"strings"
	"time"

	"github.com/0xERR0R/blocky/util"
	"github.com/miekg/dns"
	"golang.org/x/net/publicsuffix"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type logEntry struct {
	RequestTS     *time.Time `gorm:"index"`
	ClientIP      string
	ClientName    string `gorm:"index"`
	DurationMs    int64
	Reason        string
	ResponseType  string `gorm:"index"`
	QuestionType  string
	QuestionName  string
	EffectiveTLDP string
	Answer        string
	ResponseCode  string
}

type DatabaseWriter struct {
	db *gorm.DB
}

func NewDatabaseWriter(target string) *DatabaseWriter {
	db, err := gorm.Open(mysql.Open(target), &gorm.Config{})

	if err != nil {
		util.FatalOnError("can't create database connection", err)
		return nil
	}

	// Migrate the schema
	err = db.AutoMigrate(&logEntry{})
	if err != nil {
		util.FatalOnError("can't perform auto migration", err)
		return nil
	}

	return &DatabaseWriter{db: db}
}

func (d *DatabaseWriter) Write(entry *Entry) {
	domain := util.ExtractDomain(entry.Request.Req.Question[0])
	etld, _ := publicsuffix.EffectiveTLDPlusOne(domain)

	d.db.Create(&logEntry{
		RequestTS:     &entry.Start,
		ClientIP:      entry.Request.ClientIP.String(),
		ClientName:    strings.Join(entry.Request.ClientNames, "; "),
		DurationMs:    entry.DurationMs,
		Reason:        entry.Response.Reason,
		ResponseType:  entry.Response.RType.String(),
		QuestionType:  dns.TypeToString[entry.Request.Req.Question[0].Qtype],
		QuestionName:  domain,
		EffectiveTLDP: etld,
		Answer:        util.AnswerToString(entry.Response.Res.Answer),
		ResponseCode:  dns.RcodeToString[entry.Response.Res.Rcode],
	})
}

func (d *DatabaseWriter) CleanUp() {
	// delete from xx where request_ts < xx
}
