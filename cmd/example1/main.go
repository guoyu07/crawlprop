package main


func exploreURL(url string, urlProcessor chan string) {

	log.Println("Visiting URL: ", url)

	resp, err := http.Get(url)
	if err != nil {
		log.Println("ERROR in HTTP.GET method: ", err)
		return
	}

	defer resp.Body.Close()
	z := html.NewTokenizer(resp.Body)

	for {
		tt := z.Next()
		switch {
		case tt == html.ErrorToken:
			return

		case tt == html.StartTagToken:
			t := z.Token()
			if t.Data == "a" {
				for _, a := range t.Attr {
					if a.Key == "href" {
						// if link is within baseurl
						if strings.HasPrefix(a.Val, baseurl) {	urlProcessor <- a.Val }
						//if strings.Contains(a.Val, baseurl) || strings.HasPrefix(a.Val, baseurl) {	urlProcessor <- a.Val }
						//urlProcessor <- a.Val
					}
				}
			}
		}
	}
}
