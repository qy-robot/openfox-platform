package model

// TeamProductIdentity is retained only so migration and reconciliation tools
// can read the historical account-to-platform team mapping. Runtime team
// authorization and mutation are owned by New API.
type TeamProductIdentity struct {
	ID             uint   `json:"id" gorm:"primaryKey"`
	Product        string `json:"product" gorm:"type:varchar(32);not null;uniqueIndex:idx_team_product_identity"`
	ExternalTeamID string `json:"external_team_id" gorm:"type:varchar(96);not null;uniqueIndex:idx_team_product_identity"`
	PlatformTeamID int    `json:"platform_team_id" gorm:"not null;uniqueIndex"`
	Version        int64  `json:"version" gorm:"not null;default:0"`
	ProjectionHash string `json:"projection_hash" gorm:"type:varchar(64);not null;default:''"`
	CreatedAt      int64  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      int64  `json:"updated_at" gorm:"autoUpdateTime"`
}

// AccountBillingOperation is a read-only historical record after the account
// billing bridge is retired. New API owns all wallet and team ledger writes.
type AccountBillingOperation struct {
	ID            uint   `json:"id" gorm:"primaryKey"`
	OperationID   string `json:"operation_id" gorm:"type:varchar(64);not null;uniqueIndex"`
	Action        string `json:"action" gorm:"type:varchar(32);not null"`
	PayloadHash   string `json:"payload_hash" gorm:"type:varchar(64);not null"`
	ResultJSON    string `json:"result_json" gorm:"type:text;not null"`
	ActorSubject  string `json:"actor_subject" gorm:"type:varchar(64);not null;default:''"`
	TargetSubject string `json:"target_subject" gorm:"type:varchar(64);not null;default:''"`
	Reason        string `json:"reason" gorm:"type:varchar(500);not null;default:''"`
	CreatedAt     int64  `json:"created_at" gorm:"autoCreateTime"`
}
