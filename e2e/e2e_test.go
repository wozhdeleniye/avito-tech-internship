package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"testing"
	"time"
)

func baseURL() string {
	if v := os.Getenv("SERVER_PORT"); v != "" {

		return "http://localhost:" + v
	}
	return "http://localhost:8085"
}

func postJSON(t *testing.T, path string, body interface{}) (*http.Response, []byte) {
	t.Helper()
	b, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("ошибка сериализации тела: %v", err)
	}
	url := baseURL() + path
	res, err := http.Post(url, "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("POST %s не удался: %v", url, err)
	}
	data, _ := io.ReadAll(res.Body)
	res.Body.Close()
	return res, data
}

func get(t *testing.T, path string) (*http.Response, []byte) {
	t.Helper()
	url := baseURL() + path
	res, err := http.Get(url)
	if err != nil {
		t.Fatalf("GET %s не удался: %v", url, err)
	}
	data, _ := io.ReadAll(res.Body)
	res.Body.Close()
	return res, data
}

func uniqueName(prefix string) string {
	return fmt.Sprintf("%s-%d-%d", prefix, time.Now().UnixNano(), rand.Intn(1000))
}

func TestCreateTeamAndGet(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	team := uniqueName("e2e-team")
	member := map[string]interface{}{"user_id": team + "-u1", "username": "u1", "is_active": true}
	body := map[string]interface{}{"team_name": team, "members": []interface{}{member}}

	res, data := postJSON(t, "/api/team/add", body)
	if res.StatusCode != http.StatusCreated && res.StatusCode != http.StatusOK {
		t.Fatalf("ожидался статус 201/200, получено %d: %s", res.StatusCode, string(data))
	}

	res, data = get(t, "/api/team/get?team_name="+team)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("GET команды ожидался 200, получено %d: %s", res.StatusCode, string(data))
	}
}

func TestCreatePRAndAssign(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	team := uniqueName("e2e-team-pr")
	members := []interface{}{}
	for i := 0; i < 3; i++ {
		members = append(members, map[string]interface{}{"user_id": fmt.Sprintf("%s-u%d", team, i), "username": fmt.Sprintf("u%d", i), "is_active": true})
	}
	body := map[string]interface{}{"team_name": team, "members": members}
	res, data := postJSON(t, "/api/team/add", body)
	if res.StatusCode != http.StatusCreated && res.StatusCode != http.StatusOK {
		t.Fatalf("создание команды не удалось: %d %s", res.StatusCode, string(data))
	}

	author := fmt.Sprintf("%s-u0", team)
	prBody := map[string]interface{}{"author_id": author, "pull_request_id": uniqueName("pr"), "pull_request_name": "e2e-pr"}
	res, data = postJSON(t, "/api/pullRequest/create", prBody)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("создание PR ожидалось 201, получено %d: %s", res.StatusCode, string(data))
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(data, &resp); err != nil {
		t.Fatalf("ошибка разбора ответа при создании PR: %v; тело: %s", err, string(data))
	}
}

func createTeam(t *testing.T, team string, n int) []string {
	members := make([]interface{}, 0, n)
	ids := make([]string, 0, n)
	for i := 0; i < n; i++ {
		uid := fmt.Sprintf("%s-u%d", team, i)
		members = append(members, map[string]interface{}{"user_id": uid, "username": fmt.Sprintf("u%d", i), "is_active": true})
		ids = append(ids, uid)
	}
	body := map[string]interface{}{"team_name": team, "members": members}
	res, data := postJSON(t, "/api/team/add", body)
	if res.StatusCode != http.StatusCreated && res.StatusCode != http.StatusOK {
		t.Fatalf("создание команды не удалось: %d %s", res.StatusCode, string(data))
	}
	return ids
}

func createPR(t *testing.T, author string) (string, []string) {
	prID := uniqueName("pr")
	prBody := map[string]interface{}{"author_id": author, "pull_request_id": prID, "pull_request_name": "e2e-pr"}
	res, data := postJSON(t, "/api/pullRequest/create", prBody)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("create pr expected 201 got %d: %s", res.StatusCode, string(data))
	}
	var body map[string]map[string]interface{}
	if err := json.Unmarshal(data, &body); err != nil {
		t.Fatalf("ошибка разбора ответа при создании PR: %v; тело: %s", err, string(data))
	}
	prObj, ok := body["pr"]
	if !ok {
		t.Fatalf("ответ не содержит поле 'pr': %s", string(data))
	}
	assigned := []string{}
	if ar, ok := prObj["assigned_reviewers"].([]interface{}); ok {
		for _, v := range ar {
			if s, ok := v.(string); ok {
				assigned = append(assigned, s)
			}
		}
	}
	return prID, assigned
}

func TestMergePullRequest(t *testing.T) {
	team := uniqueName("e2e-merge-team")
	ids := createTeam(t, team, 3)
	author := ids[0]
	prID, _ := createPR(t, author)

	res, data := postJSON(t, "/api/pullRequest/merge", map[string]interface{}{"pull_request_id": prID})
	if res.StatusCode != http.StatusOK {
		t.Fatalf("слияние ожидалось 200, получено %d: %s", res.StatusCode, string(data))
	}
	var b map[string]map[string]interface{}
	if err := json.Unmarshal(data, &b); err == nil {
		if pr, ok := b["pr"]; ok {
			if status, ok := pr["status"].(string); ok {
				if status != "MERGED" {
					t.Fatalf("ожидался статус MERGED, получен %s", status)
				}
			}
		}
	}
}

func TestReassignReviewerAndGetReview(t *testing.T) {
	team := uniqueName("e2e-reassign-team")
	ids := createTeam(t, team, 6)
	author := ids[0]
	prID, assigned := createPR(t, author)
	if len(assigned) == 0 {
		t.Fatalf("нет назначенных ревьюверов для PR %s", prID)
	}
	oldReviewer := assigned[0]

	res, data := postJSON(t, "/api/pullRequest/reassign", map[string]interface{}{"pull_request_id": prID, "old_user_id": oldReviewer})
	if res.StatusCode != http.StatusOK {
		t.Fatalf("переназначение ожидалось 200, получено %d: %s", res.StatusCode, string(data))
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(data, &resp); err != nil {
		t.Fatalf("ошибка разбора ответа переназначения: %v; тело: %s", err, string(data))
	}
	if _, ok := resp["new_reviewer_id"]; !ok {
	}

	checkUser := oldReviewer
	res2, data2 := get(t, "/api/users/getReview?user_id="+checkUser)
	if res2.StatusCode != http.StatusOK {
		t.Fatalf("getReview ожидался 200, получено %d: %s", res2.StatusCode, string(data2))
	}
}

func TestSetUserIsActive(t *testing.T) {
	team := uniqueName("e2e-setactive-team")
	ids := createTeam(t, team, 3)
	user := ids[1]

	res, data := postJSON(t, "/api/users/setIsActive", map[string]interface{}{"user_id": user, "is_active": false})
	if res.StatusCode != http.StatusOK {
		t.Fatalf("setIsActive ожидался 200, получено %d: %s", res.StatusCode, string(data))
	}

	res2, data2 := get(t, "/api/team/get?team_name="+team)
	if res2.StatusCode != http.StatusOK {
		t.Fatalf("GET команды ожидался 200, получено %d: %s", res2.StatusCode, string(data2))
	}
	var teamResp map[string]interface{}
	if err := json.Unmarshal(data2, &teamResp); err != nil {
		t.Fatalf("ошибка разбора команды: %v", err)
	}
	if members, ok := teamResp["members"].([]interface{}); ok {
		found := false
		for _, m := range members {
			if mm, ok := m.(map[string]interface{}); ok {
				if mm["user_id"] == user {
					found = true
					if active, ok := mm["is_active"].(bool); ok {
						if active {
							t.Fatalf("ожидалось, что пользователь %s будет неактивен", user)
						}
					}
				}
			}
		}
		if !found {
			t.Fatalf("пользователь %s не найден среди участников команды", user)
		}
	}
}

func TestAdminMassDeactivateAndReassign(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	oldTeam := uniqueName("e2e-old")
	newTeam := uniqueName("e2e-new")

	oldMembers := []interface{}{}
	for i := 0; i < 3; i++ {
		oldMembers = append(oldMembers, map[string]interface{}{"user_id": fmt.Sprintf("%s-u%d", oldTeam, i), "username": fmt.Sprintf("o%d", i), "is_active": true})
	}
	_, data := postJSON(t, "/api/team/add", map[string]interface{}{"team_name": oldTeam, "members": oldMembers})
	_ = data

	newMembers := []interface{}{}
	for i := 0; i < 3; i++ {
		newMembers = append(newMembers, map[string]interface{}{"user_id": fmt.Sprintf("%s-u%d", newTeam, i), "username": fmt.Sprintf("n%d", i), "is_active": true})
	}
	_, _ = postJSON(t, "/api/team/add", map[string]interface{}{"team_name": newTeam, "members": newMembers})

	author := fmt.Sprintf("%s-u0", oldTeam)
	prID := uniqueName("e2e-pr")
	prBody := map[string]interface{}{"author_id": author, "pull_request_id": prID, "pull_request_name": "e2e-pr"}
	res, data := postJSON(t, "/api/pullRequest/create", prBody)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("создание PR ожидалось 201, получено %d: %s", res.StatusCode, string(data))
	}

	req := map[string]interface{}{"old_team_name": oldTeam, "new_team_name": newTeam}
	res, data = postJSON(t, "/api/admin/team/deactivate", req)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("admin deactivate ожидался 200, получено %d: %s", res.StatusCode, string(data))
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(data, &resp); err != nil {
		t.Fatalf("ошибка разбора ответа admin: %v; тело: %s", err, string(data))
	}
	if _, ok := resp["deactivated"]; !ok {
		t.Fatalf("ответ admin не содержит 'deactivated': %s", string(data))
	}
}
