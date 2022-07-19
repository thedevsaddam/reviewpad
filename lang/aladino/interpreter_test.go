// Copyright 2022 Explore.dev Unipessoal Lda. All Rights Reserved.
// Use of this source code is governed by a license that can be
// found in the LICENSE file.

package aladino_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"testing"

	"github.com/google/go-github/v42/github"
	"github.com/migueleliasweb/go-github-mock/src/mock"
	"github.com/reviewpad/reviewpad/v2/engine"
	"github.com/reviewpad/reviewpad/v2/lang/aladino"
	mocks_aladino "github.com/reviewpad/reviewpad/v2/mocks/aladino"
	"github.com/stretchr/testify/assert"
)

func TestProcessGroupWhenGroupTypeFilterIsSetAndBuildGroupASTFails(t *testing.T) {
	mockedEnv, err := mocks_aladino.MockDefaultEnv(nil, nil)
	if err != nil {
		log.Fatalf("mockDefaultEnv failed: %v", err)
	}

	mockedInterpreter := &aladino.Interpreter{
		Env: mockedEnv,
	}

	groupName := "senior-developers"

	err = mockedInterpreter.ProcessGroup(
		groupName,
		engine.GroupKindDeveloper,
		engine.GroupTypeFilter,
		"$group(\""+groupName+"\")",
		"dev",
		"$hasFileExtensions(",
	)

	assert.EqualError(t, err, "buildGroupAST: parse error: failed to build AST on input $hasFileExtensions(")
}

func TestProcessGroupWhenGroupTypeFilterIsSet(t *testing.T) {
	member := "jane"
	ghMembers := []*github.User{
		{Login: github.String(member)},
	}
	mockedPullRequestFileList := &[]*github.CommitFile{
		{
			Filename: github.String("default-mock-repo/file1.ts"),
			Patch:    nil,
		},
	}
	mockedEnv, err := mocks_aladino.MockDefaultEnv(
		[]mock.MockBackendOption{
			mock.WithRequestMatch(
				mock.GetOrgsMembersByOrg,
				ghMembers,
			),
			mock.WithRequestMatchHandler(
				mock.GetReposPullsFilesByOwnerByRepoByPullNumber,
				http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.Write(mock.MustMarshal(mockedPullRequestFileList))
				}),
			),
		},
		nil,
	)
	if err != nil {
		log.Fatalf("mockDefaultEnv failed: %v", err)
	}

	mockedInterpreter := &aladino.Interpreter{
		Env: mockedEnv,
	}

	groupName := "senior-developers"

	err = mockedInterpreter.ProcessGroup(
		groupName,
		engine.GroupKindDeveloper,
		engine.GroupTypeFilter,
		"$group(\""+groupName+"\")",
		"dev",
		"$hasFileExtensions([\".ts\"])",
	)

	gotVal := mockedEnv.GetRegisterMap()[groupName]

	wantVal := aladino.BuildArrayValue([]aladino.Value{
		aladino.BuildStringValue(member),
	})

	assert.Nil(t, err)
	assert.Equal(t, wantVal, gotVal)
}

func TestProcessGroupWhenGroupTypeFilterIsNotSet(t *testing.T) {
	mockedEnv, err := mocks_aladino.MockDefaultEnv(nil, nil)
	if err != nil {
		log.Fatalf("mockDefaultEnv failed: %v", err)
	}

	mockedInterpreter := &aladino.Interpreter{
		Env: mockedEnv,
	}

	groupName := "senior-developers"

	err = mockedInterpreter.ProcessGroup(
		groupName,
		engine.GroupKindDeveloper,
		engine.GroupTypeStatic,
		"[\"jane\"]",
		"",
		"",
	)

	gotVal := mockedEnv.GetRegisterMap()[groupName]

	wantVal := aladino.BuildArrayValue([]aladino.Value{
		aladino.BuildStringValue("jane"),
	})

	assert.Nil(t, err)
	assert.Equal(t, wantVal, gotVal)
}

func TestProcessGroupWhenGroupTypeFilterIsNotSetAndEvalGroupTypeInferenceFails(t *testing.T) {
	mockedEnv, err := mocks_aladino.MockDefaultEnv(nil, nil)
	if err != nil {
		log.Fatalf("mockDefaultEnv failed: %v", err)
	}

	mockedInterpreter := &aladino.Interpreter{
		Env: mockedEnv,
	}

	groupName := "senior-developers"

	err = mockedInterpreter.ProcessGroup(
		groupName,
		engine.GroupKindDeveloper,
		engine.GroupTypeStatic,
		"$group(1)",
		"",
		"",
	)

	assert.EqualError(t, err, "type inference failed: mismatch in arg types on group")
}

func TestProcessGroupWhenGroupTypeFilterIsNotSetAndGroupExprIsNotAnArray(t *testing.T) {
	mockedEnv, err := mocks_aladino.MockDefaultEnv(nil, nil)
	if err != nil {
		log.Fatalf("mockDefaultEnv failed: %v", err)
	}

	mockedInterpreter := &aladino.Interpreter{
		Env: mockedEnv,
	}

	groupName := "senior-developers"

	err = mockedInterpreter.ProcessGroup(
		groupName,
		engine.GroupKindDeveloper,
		engine.GroupTypeStatic,
		"1",
		"",
		"",
	)

	assert.EqualError(t, err, "expression is not a valid group")
}

func TestBuildInternalLabelID(t *testing.T) {
	labelID := "label_id"

	wantVal := fmt.Sprintf("@label:%v", labelID)

	gotVal := aladino.BuildInternalLabelID(labelID)

	assert.Equal(t, wantVal, gotVal)
}

func TestProcessLabel(t *testing.T) {
	mockedEnv, err := mocks_aladino.MockDefaultEnv(nil, nil)
	if err != nil {
		log.Fatalf("mockDefaultEnv failed: %v", err)
	}

	mockedInterpreter := &aladino.Interpreter{
		Env: mockedEnv,
	}

	labelID := "label_id"
	labelName := "label_name"
	err = mockedInterpreter.ProcessLabel(labelID, labelName)

	internalLabelID := fmt.Sprintf("@label:%v", labelID)
	gotVal := mockedEnv.GetRegisterMap()[internalLabelID]

	wantVal := aladino.BuildStringValue(labelName)

	assert.Nil(t, err)
	assert.Equal(t, wantVal, gotVal)
}

func TestBuildInternalRuleName(t *testing.T) {
	ruleName := "rule_name"

	wantVal := fmt.Sprintf("@rule:%v", ruleName)

	gotVal := aladino.BuildInternalRuleName(ruleName)

	assert.Equal(t, wantVal, gotVal)
}

func TestProcessRule(t *testing.T) {
	mockedEnv, err := mocks_aladino.MockDefaultEnv(nil, nil)
	if err != nil {
		log.Fatalf("mockDefaultEnv failed: %v", err)
	}

	mockedInterpreter := &aladino.Interpreter{
		Env: mockedEnv,
	}

	ruleName := "rule_name"
	spec := "1 == 1"
	err = mockedInterpreter.ProcessRule(ruleName, spec)

	internalRuleName := fmt.Sprintf("@rule:%v", ruleName)
	gotVal := mockedEnv.GetRegisterMap()[internalRuleName]

	wantVal := aladino.BuildStringValue(spec)

	assert.Nil(t, err)
	assert.Equal(t, wantVal, gotVal)
}

func TestEvalExpr_WhenParseFails(t *testing.T) {
	mockedEnv, err := mocks_aladino.MockDefaultEnv(nil, nil)
	if err != nil {
		log.Fatalf("mockDefaultEnv failed: %v", err)
	}

	gotVal, err := aladino.EvalExpr(mockedEnv, "", "1 ==")

	assert.False(t, gotVal)
	assert.EqualError(t, err, "parse error: failed to build AST on input 1 ==")
}

func TestEvalExpr_WhenTypeInferenceFails(t *testing.T) {
	mockedEnv, err := mocks_aladino.MockDefaultEnv(nil, nil)
	if err != nil {
		log.Fatalf("mockDefaultEnv failed: %v", err)
	}

	gotVal, err := aladino.EvalExpr(mockedEnv, "", "1 == \"a\"")

	assert.False(t, gotVal)
	assert.EqualError(t, err, "type inference failed")
}

func TestEvalExpr_WhenExprIsNotBoolType(t *testing.T) {
	mockedEnv, err := mocks_aladino.MockDefaultEnv(nil, nil)
	if err != nil {
		log.Fatalf("mockDefaultEnv failed: %v", err)
	}

	gotVal, err := aladino.EvalExpr(mockedEnv, "", "1")

	assert.False(t, gotVal)
	assert.EqualError(t, err, "expression 1 is not a condition")
}

func TestEvalExpr(t *testing.T) {
	mockedEnv, err := mocks_aladino.MockDefaultEnv(nil, nil)
	if err != nil {
		log.Fatalf("mockDefaultEnv failed: %v", err)
	}

	gotVal, err := aladino.EvalExpr(mockedEnv, "", "1 == 1")

	assert.Nil(t, err)
	assert.True(t, gotVal)
}

func TestEvalExpr_OnInterpreter(t *testing.T) {
	mockedEnv, err := mocks_aladino.MockDefaultEnv(nil, nil)
	if err != nil {
		log.Fatalf("mockDefaultEnv failed: %v", err)
	}

	mockedInterpreter := &aladino.Interpreter{
		Env: mockedEnv,
	}

	gotVal, err := mockedInterpreter.EvalExpr("", "1 == 1")

	assert.Nil(t, err)
	assert.True(t, gotVal)
}

func TestExecProgram_WhenExecStatementFails(t *testing.T) {
	mockedEnv, err := mocks_aladino.MockDefaultEnv(nil, nil)
	if err != nil {
		log.Fatalf("mockDefaultEnv failed: %v", err)
	}

	mockedInterpreter := &aladino.Interpreter{
		Env: mockedEnv,
	}

	program := &engine.Program{
		Statements: []*engine.Statement{
			{
				Code:     "$action()",
				Metadata: nil,
			},
		},
	}

	err = mockedInterpreter.ExecProgram(program)

	assert.EqualError(t, err, "no type for built-in action. Please check if the mode in the reviewpad.yml file supports it.")
}

func TestExecProgram(t *testing.T) {
	mockedEnv, err := mocks_aladino.MockDefaultEnv(
		[]mock.MockBackendOption{
			mock.WithRequestMatch(
				mock.GetReposLabelsByOwnerByRepoByName,
				&github.Label{},
			),
			mock.WithRequestMatch(
				mock.PostReposIssuesLabelsByOwnerByRepoByIssueNumber,
				[]*github.Label{
					{Name: github.String("test")},
				},
			),
		},
		nil,
	)
	if err != nil {
		log.Fatalf("mockDefaultEnv failed: %v", err)
	}

	mockedInterpreter := &aladino.Interpreter{
		Env: mockedEnv,
	}

	statementWorkflowName := "test"
	statementRule := "testRule"
	statementCode := "$addLabel(\"test\")"
	statement := &engine.Statement{
		Code: statementCode,
		Metadata: &engine.Metadata{
			Workflow: engine.PadWorkflow{
				Name: statementWorkflowName,
			},
			TriggeredBy: []engine.PadWorkflowRule{
				{Rule: statementRule},
			},
		},
	}

	program := &engine.Program{
		Statements: []*engine.Statement{
			statement,
		},
	}

	err = mockedInterpreter.ExecProgram(program)

	gotVal := mockedEnv.GetReport().WorkflowDetails[statementWorkflowName]

	wantVal := aladino.ReportWorkflowDetails{
		Name: statementWorkflowName,
		Rules: map[string]bool{
			statementRule: true,
		},
		Actions: []string{statementCode},
	}

	assert.Nil(t, err)
	assert.Equal(t, wantVal, gotVal)
}

func TestExecStatement_WhenParseFails(t *testing.T) {
	mockedEnv, err := mocks_aladino.MockDefaultEnv(nil, nil)
	if err != nil {
		log.Fatalf("mockDefaultEnv failed: %v", err)
	}

	mockedInterpreter := &aladino.Interpreter{
		Env: mockedEnv,
	}

	statement := &engine.Statement{
		Code: "$addLabel(",
		Metadata: &engine.Metadata{
			Workflow: engine.PadWorkflow{
				Name: "test",
			},
			TriggeredBy: []engine.PadWorkflowRule{
				{Rule: "testRule"},
			},
		},
	}

	err = mockedInterpreter.ExecStatement(statement)

	assert.EqualError(t, err, "parse error: failed to build AST on input $addLabel(")
}

func TestExecStatement_WhenTypeCheckExecFails(t *testing.T) {
	mockedEnv, err := mocks_aladino.MockDefaultEnv(nil, nil)
	if err != nil {
		log.Fatalf("mockDefaultEnv failed: %v", err)
	}

	mockedInterpreter := &aladino.Interpreter{
		Env: mockedEnv,
	}

	statement := &engine.Statement{
		Code: "$addLabel(1)",
		Metadata: &engine.Metadata{
			Workflow: engine.PadWorkflow{
				Name: "test",
			},
			TriggeredBy: []engine.PadWorkflowRule{
				{Rule: "testRule"},
			},
		},
	}

	err = mockedInterpreter.ExecStatement(statement)

	assert.EqualError(t, err, "type inference failed: mismatch in arg types on addLabel")
}

func TestExecStatement_WhenActionExecFails(t *testing.T) {
	mockedEnv, err := mocks_aladino.MockDefaultEnv(nil, nil)
	if err != nil {
		log.Fatalf("mockDefaultEnv failed: %v", err)
	}

	mockedInterpreter := &aladino.Interpreter{
		Env: mockedEnv,
	}

	statement := &engine.Statement{
		Code: "$author()",
		Metadata: &engine.Metadata{
			Workflow: engine.PadWorkflow{
				Name: "test",
			},
			TriggeredBy: []engine.PadWorkflowRule{
				{Rule: "testRule"},
			},
		},
	}

	err = mockedInterpreter.ExecStatement(statement)

	assert.EqualError(t, err, "exec: author not found. are you sure this is a built-in function?")
}

func TestExecStatement(t *testing.T) {
	mockedEnv, err := mocks_aladino.MockDefaultEnv(
		[]mock.MockBackendOption{
			mock.WithRequestMatch(
				mock.GetReposLabelsByOwnerByRepoByName,
				&github.Label{},
			),
			mock.WithRequestMatch(
				mock.PostReposIssuesLabelsByOwnerByRepoByIssueNumber,
				[]*github.Label{
					{Name: github.String("test")},
				},
			),
		},
		nil,
	)
	if err != nil {
		log.Fatalf("mockDefaultEnv failed: %v", err)
	}

	mockedInterpreter := &aladino.Interpreter{
		Env: mockedEnv,
	}

	statementWorkflowName := "test"
	statementRule := "testRule"
	statementCode := "$addLabel(\"test\")"
	statement := &engine.Statement{
		Code: statementCode,
		Metadata: &engine.Metadata{
			Workflow: engine.PadWorkflow{
				Name: statementWorkflowName,
			},
			TriggeredBy: []engine.PadWorkflowRule{
				{Rule: statementRule},
			},
		},
	}

	err = mockedInterpreter.ExecStatement(statement)

	gotVal := mockedEnv.GetReport().WorkflowDetails[statementWorkflowName]

	wantVal := aladino.ReportWorkflowDetails{
		Name: statementWorkflowName,
		Rules: map[string]bool{
			statementRule: true,
		},
		Actions: []string{statementCode},
	}

	assert.Nil(t, err)
	assert.Equal(t, wantVal, gotVal)
}

func TestReport_WhenFindReportCommentFails(t *testing.T) {
	mockedPullRequest := mocks_aladino.GetDefaultMockPullRequestDetailsWith(&github.PullRequest{
		User: &github.User{Login: github.String("john")},
		Base: &github.PullRequestBranch{
			Repo: &github.Repository{
				Owner: &github.User{
					Login: github.String("john"),
				},
				Name: github.String("default-mock-repo"),
			},
			Ref: github.String("master"),
		},
	})
	mockedEnv, err := mocks_aladino.MockDefaultEnv(
		[]mock.MockBackendOption{
			mock.WithRequestMatchHandler(
				mock.GetReposPullsByOwnerByRepoByPullNumber,
				http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.Write(mock.MustMarshal(mockedPullRequest))
				}),
			),
		},
		nil,
	)
	if err != nil {
		log.Fatalf("mockDefaultEnv failed: %v", err)
	}

	mockedInterpreter := &aladino.Interpreter{
		Env: mockedEnv,
	}

	err = mockedInterpreter.Report(engine.SILENT_MODE)

	assert.EqualError(t, err, "[report] error getting issues mock response not found for /repos/john/default-mock-repo/issues/6/comments")
}

func TestReport_OnSilentMode_WhenThereIsAlreadyAReviewpadComment(t *testing.T) {
	var isDeletedCommentRequested bool
	mockedEnv, err := mocks_aladino.MockDefaultEnv(
		[]mock.MockBackendOption{
			mock.WithRequestMatch(
				mock.GetReposIssuesCommentsByOwnerByRepoByIssueNumber,
				[]*github.IssueComment{
                    {
                        ID: github.Int64(1234),
                        Body: github.String("<!--@annotation-reviewpad-report-->\n**Reviewpad Report**\n\n:scroll: **Explanation**\nNo workflows activated"),
                    },
                },
			),
			mock.WithRequestMatchHandler(
				mock.DeleteReposIssuesCommentsByOwnerByRepoByCommentId,
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					// When a pull request has a reviewpad comment then this comment is deleted
					isDeletedCommentRequested = true
				}),
			),
		},
		nil,
	)
	if err != nil {
		log.Fatalf("mockDefaultEnv failed: %v", err)
	}

	mockedInterpreter := &aladino.Interpreter{
		Env: mockedEnv,
	}

	err = mockedInterpreter.Report(engine.SILENT_MODE)

	assert.Nil(t, err)
	assert.True(t, isDeletedCommentRequested)
}

func TestReport_OnSilentMode_WhenNoReviewpadCommentIsFound(t *testing.T) {
	var isDeletedCommentRequested bool
	mockedEnv, err := mocks_aladino.MockDefaultEnv(
		[]mock.MockBackendOption{
			mock.WithRequestMatch(
				mock.GetReposIssuesCommentsByOwnerByRepoByIssueNumber,
				[]*github.IssueComment{},
			),
			mock.WithRequestMatchHandler(
				mock.DeleteReposIssuesCommentsByOwnerByRepoByCommentId,
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					// When a pull request has a reviewpad comment then this comment is deleted
					isDeletedCommentRequested = true
				}),
			),
		},
		nil,
	)
	if err != nil {
		log.Fatalf("mockDefaultEnv failed: %v", err)
	}

	mockedInterpreter := &aladino.Interpreter{
		Env: mockedEnv,
	}

	err = mockedInterpreter.Report(engine.SILENT_MODE)

	assert.Nil(t, err)
	assert.False(t, isDeletedCommentRequested)
}

func TestReport_OnVerboseMode_WhenNoReviewpadCommentIsFound(t *testing.T) {
	var addedComment string
	commentToBeAdded := "<!--@annotation-reviewpad-report-->\n**Reviewpad Report**\n\n:scroll: **Explanation**\nNo workflows activated"
	mockedEnv, err := mocks_aladino.MockDefaultEnv(
		[]mock.MockBackendOption{
			mock.WithRequestMatch(
				mock.GetReposIssuesCommentsByOwnerByRepoByIssueNumber,
				[]*github.IssueComment{},
			),
			mock.WithRequestMatchHandler(
				mock.PostReposIssuesCommentsByOwnerByRepoByIssueNumber,
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					rawBody, _ := ioutil.ReadAll(r.Body)
					body := github.IssueComment{}

					json.Unmarshal(rawBody, &body)

					addedComment = *body.Body
				}),
			),
		},
		nil,
	)
	if err != nil {
		log.Fatalf("mockDefaultEnv failed: %v", err)
	}

	mockedInterpreter := &aladino.Interpreter{
		Env: mockedEnv,
	}

	err = mockedInterpreter.Report(engine.VERBOSE_MODE)

	assert.Nil(t, err)
	assert.Equal(t, commentToBeAdded, addedComment)
}

func TestReport_OnVerboseMode_WhenThereIsAlreadyAReviewpadComment(t *testing.T) {
	var updatedComment string
	commentUpdated := "<!--@annotation-reviewpad-report-->\n**Reviewpad Report**\n\n:scroll: **Explanation**\nNo workflows activated"
	mockedEnv, err := mocks_aladino.MockDefaultEnv(
		[]mock.MockBackendOption{
			mock.WithRequestMatch(
				mock.GetReposIssuesCommentsByOwnerByRepoByIssueNumber,
				[]*github.IssueComment{
					{
						ID:   github.Int64(1234),
						Body: github.String("<!--@annotation-reviewpad-report-->\n**Reviewpad Report**\n**:information_source: Messages**\n* Changes the README.md"),
					},
				},
			),
			mock.WithRequestMatchHandler(
				mock.PatchReposIssuesCommentsByOwnerByRepoByCommentId,
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					rawBody, _ := ioutil.ReadAll(r.Body)
					body := github.IssueComment{}

					json.Unmarshal(rawBody, &body)

					updatedComment = *body.Body
				}),
			),
		},
		nil,
	)
	if err != nil {
		log.Fatalf("mockDefaultEnv failed: %v", err)
	}

	mockedInterpreter := &aladino.Interpreter{
		Env: mockedEnv,
	}

	err = mockedInterpreter.Report(engine.VERBOSE_MODE)

	assert.Nil(t, err)
	assert.Equal(t, commentUpdated, updatedComment)
}

func TestNewInterpreter_WhenNewEvalEnvFails(t *testing.T) {
	ctx := context.Background()
	failMessage := "GetPullRequestFilesRequestFail"
	client := github.NewClient(
		mock.NewMockedHTTPClient(
			mock.WithRequestMatchHandler(
				mock.GetReposPullsFilesByOwnerByRepoByPullNumber,
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					mock.WriteError(
						w,
						http.StatusInternalServerError,
						failMessage,
					)
				}),
			),
		),
	)
	gotInterpreter, err := aladino.NewInterpreter(
		ctx,
		client,
		nil,
		nil,
		mocks_aladino.GetDefaultMockPullRequestDetails(),
		nil,
	)

	assert.Nil(t, gotInterpreter)
	assert.NotNil(t, err)
}

func TestNewInterpreter(t *testing.T) {
	mockedEnv, err := mocks_aladino.MockDefaultEnv(nil, nil)
	if err != nil {
		log.Fatalf("mockDefaultEnv failed: %v", err)
	}

	wantInterpreter := &aladino.Interpreter{
		Env: mockedEnv,
	}

	gotInterpreter, err := aladino.NewInterpreter(
		mockedEnv.GetCtx(),
		mockedEnv.GetClient(),
		mockedEnv.GetClientGQL(),
		mockedEnv.GetCollector(),
		mockedEnv.GetPullRequest(),
		mockedEnv.GetBuiltIns(),
	)

	assert.Nil(t, err)
	assert.Equal(t, wantInterpreter, gotInterpreter)
}
