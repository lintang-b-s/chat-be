package entity

type AddFriendRequest struct {
	MyUsername     string `json:"my_username"`
	FriendUsername string `json:"friend_username"`
}

type GetContactRequest struct {
	MyUsername string `json:"my_username"`
}
