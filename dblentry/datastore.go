package dblentry

import "fmt"
import "strings"

type Datastore struct {
	name    string
	transdb *DB
	pricedb *DB
	accntdb map[string]*Account // full account-name -> account
	// directive fields
	year         int               // year
	month        int               // month
	dateformat   string            // dataformat
	aliases      map[string]string // alias, account-alias
	payees       map[string]string // account-payee
	rootaccount  string            // apply-account
	blncingaccnt string            // account
}

func NewDatastore(name string) *Datastore {
	db := &Datastore{
		name:    name,
		transdb: NewDB(fmt.Sprintf("%v-transactions", name)),
		pricedb: NewDB(fmt.Sprintf("%v-pricedb", name)),
		accntdb: map[string]*Account{},
		// directives
		year:       -1,
		month:      -1,
		dateformat: "%Y/%m/%d %h:%n:%s", // TODO: no magic string
		aliases:    map[string]string{},
	}
	db.defaultprices()
	return db
}

func (db *Datastore) GetAccount(name string) *Account {
	names := strings.Split(name, ":")
	fullname := ""
	for _, name := range names {
		fullname = strings.Join([]string{fullname, name}, ":")
		if _, ok := db.accntdb[fullname]; ok == false {
			db.accntdb[fullname] = NewAccount(fullname)
		}
	}
	return db.accntdb[name]
}

func (db *Datastore) SubAccounts(parentname string) []*Account {
	accounts := []*Account{}
	for name, account := range db.accntdb {
		if strings.HasPrefix(parentname, name) {
			accounts = append(accounts, account)
		}
	}
	return accounts
}

func (db *Datastore) Apply(obj interface{}) error {
	switch blk := obj.(type) {
	case *Transaction:
		return db.transdb.Insert(blk.date, blk)

	case *Price:
		return db.pricedb.Insert(blk.when, blk)

	case *Directive:
		switch blk.dtype {
		case "year":
			db.SetYear(blk.year)
		case "month":
			db.SetMonth(blk.month)
		case "dateformat":
			db.SetDateformat(blk.dateformat)
		case "account":
			db.Declare(blk.account) // NOTE: this is redundant
		case "apply":
			db.rootaccount = blk.account.name
		case "alias":
			db.AddAlias(blk.aliasname, blk.account.name)
		case "assert":
			return fmt.Errorf("directive not-implemented")
		default:
			panic("unreachable code")
		}

	default:
		panic("unreachable code")
	}
	return nil
}

// directive-year

func (db *Datastore) SetYear(year int) *Datastore {
	db.year = year
	return db
}

func (db *Datastore) Year() int {
	return db.year
}

// directive-month

func (db *Datastore) SetMonth(month int) *Datastore {
	db.month = month
	return db
}

func (db *Datastore) Month() int {
	return db.month
}

// directive-dateformat

func (db *Datastore) SetDateformat(format string) *Datastore {
	db.dateformat = format
	return db
}

func (db *Datastore) Dateformat() string {
	return db.dateformat
}

// directive-alias

func (db *Datastore) AddAlias(aliasname, accountname string) *Datastore {
	db.aliases[aliasname] = accountname
	return db
}

func (db *Datastore) GetAlias(aliasname string) (accountname string, ok bool) {
	accountname, ok = db.aliases[aliasname]
	return accountname, ok
}

// directive-apply-account

func (db *Datastore) SetRootaccount(name string) *Datastore {
	db.rootaccount = name
	return db
}

func (db *Datastore) Rootaccount() string {
	return db.rootaccount
}

// directive-account

func (db *Datastore) Declare(value interface{}) {
	switch v := value.(type) {
	case *Account:
		account := db.GetAccount(v.name)
		account.SetDirective(v)
		if v.defblns {
			db.SetBalancingaccount(v.name)
		}

	default:
		panic("unreachable code")
	}
	panic("unreachable code")
}

func (db *Datastore) AddPayee(regex, accountname string) *Datastore {
	db.payees[regex] = accountname
	return db
}

func (db *Datastore) SetBalancingaccount(name string) *Datastore {
	db.blncingaccnt = name
	return db
}

func (db *Datastore) defaultprices() {
	_ = []string{
		"P 01/01/2000 kb 1024b",
		"P 01/01/2000 mb 1024kb",
		"P 01/01/2000 gb 1024mb",
		"P 01/01/2000 tb 1024gb",
		"P 01/01/2000 pb 1024tb",

		"P 01/01/2000 m 60s",
		"P 01/01/2000 h 60m",
	}
}