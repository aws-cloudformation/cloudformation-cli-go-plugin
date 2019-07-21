package errs

type TerminalError struct {
	CustomerFacingErrorMessage string //Customer facing errorMessage

}

func (e *TerminalError) Error() string {
	return e.CustomerFacingErrorMessage
}
