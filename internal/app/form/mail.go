package form

type SendMail struct {
	ToMail  string `binding:"Required;AlphaDashDot;MaxSize(35)"`
	Subject string
	Content string
}

type MailIDs struct {
	Ids string
	Dir string
}
