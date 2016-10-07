-- +migrate Up
CREATE TABLE recordings (
  id BIGINT,
  url TEXT,
  CONSTRAINT pk_recordings PRIMARY KEY (id)
);


-- +migrate Down
DROP TABLE recordings;
