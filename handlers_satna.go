package main

import (
	"net/http"
	"regexp"
	"time"
	"unicode"
	
	"github.com/coreos/go-log/log"
	"github.com/go-chi/chi/v5"
)

const (
	secondsBetweenAttempts = 12
	satnaTimeLimit = time.Second * secondsBetweenAttempts // mezi pokusy o heslo
)

func satnaRouter() *chi.Mux {
	r := newGameRouter("satna")
	r.Get("/", auth(satnaIndexGet))
	r.Post("/", auth(satnaIndexPost))
	r.Get("/internal", auth(satnaInternalGet))
	return r
}

func satnaIndexGet(w http.ResponseWriter, r *http.Request) {
	data := getGeneralData("Satna", w, r)
	defer func() { executeTemplate(w, "satnaIndex", data) }()

	team := server.state.GetTeam(getUser(r))
	if team != nil && team.Satna.Completed {
		http.Redirect(w, r, "/internal", http.StatusSeeOther)
		return
	}
}

func satnaIndexPost(w http.ResponseWriter, r *http.Request) {
	defer http.Redirect(w, r, "/", http.StatusSeeOther)

	if err := r.ParseForm(); err != nil {
		setFlashMessage(w, r, messageError, "Cannot parse password form")
		return
	}

	server.state.Lock()
	defer server.state.Unlock()

	team := server.state.GetTeam(getUser(r))
	if time.Since(team.Satna.LastTry) < satnaTimeLimit {
		setFlashMessage(w, r, messageError, "Časový limit mezi hesly ještě neuplynul. Mezi pokusy je potřeba čekat alespoň 12 sekund.")
		return
	}

	team.Satna.Tries++
	team.Satna.LastTry = time.Now()
	defer server.state.Save()

	// J1-l-N-s1-X-z

	password := r.PostFormValue("password")
	log.Infof("[Satna - %s] Trying password '%s'", team.Login, password)

	if len(password) < 13 {
		setFlashMessage(w, r, messageError, "Tvé heslo bylo příliš slabé, i poláci by ho zvládli uhodnout. Možná zafunguje nějaké s více znaky.")
		return
	} else if len(password) > 13 {
		setFlashMessage(w, r, messageError, "Heslo bylo dostatečně silné, jenže je tak komplikované, že se ti jej nepodaří zapamatovat správně. Zkus něco kratšího.")
		return
	}

	bpassword := []byte(password)

	reLetter := regexp.MustCompile("[a-z]")
	reBigLetter := regexp.MustCompile("[A-Z]")
	reNumbers := regexp.MustCompile("[0-9]")
	reDoubleLetters := regexp.MustCompile("[a-zA-Z][a-zA-Z]")

	smallLetters := len(reLetter.FindAll(bpassword, -1))
	bigLetters := len(reBigLetter.FindAll(bpassword, -1))
	doubleLetters := len(reDoubleLetters.FindAll(bpassword, -1))
	numbers := len(reNumbers.FindAll(bpassword, -1))
	letters := smallLetters + bigLetters
	other := len(password) - letters - numbers

	if letters < 5 {
		setFlashMessage(w, r, messageError, "Tvé heslo obsahuje tak málo písmen, že jsi byl přidán na list líných Američanů. Zkus něco s více písmeny.")
		return
	}

	if numbers < 2 {
		setFlashMessage(w, r, messageError, "Tvé heslo obsahuje málo číslic. Co neumíš počítat? Zkus něco s více číslicemi.")
		return
	}
	if doubleLetters > 0 {
		setFlashMessage(w, r, messageError, "Dvě písmena vedle sebo jsou moc podezřelá. Zkus něco bez nich.")
		return
	}
	if smallLetters != bigLetters {
		setFlashMessage(w, r, messageError, "Databáze hesel je nestabilní. K stabilitě by pomohlo, aby bylo stejně velkých jako malých písmen.")
		return
	}
	if other <= numbers {
		setFlashMessage(w, r, messageError, "Tohle není kalkulačka. Heslo má moc číslic.")
		return
	}

	req := r
	// pismena nesmi byt stejna
	wasLetter := map[rune]bool{}
	for _, r := range password {
		if unicode.IsLetter(r) {
			if _, found := wasLetter[r]; found {
				setFlashMessage(w, req, messageError, "Používat v hesle dvě stejná písmena ti přijde moc nekreativní. Takže se o použití hesla ani nepokoušíš.")
				return
			}
			wasLetter[r] = true
		}
	}

	// pismena nejsou v abecednim poradi
	lastLetter := 'a'
	for _, r := range password {
		if unicode.IsLetter(r) {
			r = unicode.ToLower(r)
			if r < lastLetter {
				setFlashMessage(w, req, messageError, "Zapomněl jsi, v jakém pořadí jdou písmenka v hesle za sebou. Raději je seřad abecedně, aby se to znovu nestalo.")
				return
			}
			lastLetter = r
		}
	}

	reJesenik := regexp.MustCompile("[jesenikJESENIK]")
	reAmericani := regexp.MustCompile("[amerikaAMERIKA]")
	reAroundNumber := regexp.MustCompile("[a-zA-Z][0-9][a-zA-Z]")
	reAroundNumberSpecial := regexp.MustCompile("[^0-9a-zA-Z][0-9][^0-9a-zA-Z]")

	jesenikLetters := len(reJesenik.FindAll(bpassword, -1))
	amerikaLetters := len(reAmericani.FindAll(bpassword, -1))

	// heslo má málo znaků z Jeseníka.
	if jesenikLetters < 3 {
		setFlashMessage(w, r, messageError, "Pro to aby sis heslo pamatoval budeš potřebovat nějakou pomůcku. Zkus tam přidat nějaké znaky ze slova Jesenik.")
		return
	}
	
	// heslo má moc znaků z Ameriky
	if amerikaLetters > 0 {
		setFlashMessage(w, r, messageError, "To jseš Americkej agent? Nemůžeš v hesle používat tajné kódy nebo budeš poslán do Gulagu. Tvé heslo obsahuje příliš mnoho znaků z Ameriky.")
		return
	}

	// cislo nesmi mit z obou stran specialni znak
	if reAroundNumberSpecial.Match(bpassword) {
		setFlashMessage(w, r, messageError, "Jakmile je číslice obklopena speciálními znaky z obou stran, je to příliš kostrbaté na vyslovení. Zkus něco jiného.")
		return
	}

	// cislo nesmi mit z obou stran pismeno
	if reAroundNumber.Match(bpassword) {
		setFlashMessage(w, r, messageError, "Číslice obklopená písmenky z obou stran vypadá jako špatně napsaná rovnice. Zkus tuto část hesla změnit.")
		return
	}

	// soucet cisel v hesle musi byt prvocislem
	sum := 0
	for _, r := range password {
		if unicode.IsDigit(r) {
			sum += int(r - '0')
		}
	}
	prime_numbers := []int{2, 3, 5, 7, 11, 13, 17, 19, 23, 29, 31, 37, 41, 43, 47, 53, 59, 61, 67, 71, 73, 79, 83, 89, 97}
	isPrime := false
	for _, prime := range prime_numbers {
		if sum == prime {
			isPrime = true
			break
		}
	}
	if !isPrime {
		setFlashMessage(w, r, messageError, "Součet číslic v hesle není roven prvočíslu. Zkus něco jiného.")
		return
	}

	log.Infof("[Satna - %s] Completed", team.Login)
	// Everything completed
	team.Satna.Completed = true
	team.Satna.CompletedTime = time.Now()
	http.Redirect(w, r, "/internal", http.StatusSeeOther)
}

func satnaInternalGet(w http.ResponseWriter, r *http.Request) {
	team := server.state.GetTeam(getUser(r))
	if team == nil || !team.Satna.Completed {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	data := getGeneralData("Satna", w, r)
	executeTemplate(w, "satnaInternal", data)
}
