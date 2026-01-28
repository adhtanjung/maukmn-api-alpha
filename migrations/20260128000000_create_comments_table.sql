-- +goose Up
-- +goose StatementBegin
CREATE TABLE comments (
    comment_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    poi_id UUID REFERENCES points_of_interest(poi_id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(user_id) NOT NULL,
    content TEXT NOT NULL,
    parent_id UUID REFERENCES comments(comment_id) ON DELETE CASCADE, -- For nested replies
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_comments_poi ON comments(poi_id);
CREATE INDEX idx_comments_user ON comments(user_id);
CREATE INDEX idx_comments_parent ON comments(parent_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS comments;
-- +goose StatementEnd
