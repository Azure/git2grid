package actions

func (as *ActionSuite) Test_Transform_Listen() {
	expected := "PullRequestEvent"
	actual := FormatEventName("pull//_request")
	as.Equal(expected, actual, "not sure what goes here")
	//as.Fail("Not Implemented!")
}
