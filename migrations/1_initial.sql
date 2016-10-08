-- +migrate Up
CREATE TABLE recordings (
  id NUMERIC,
  url TEXT,
  CONSTRAINT pk_recordings PRIMARY KEY (id)
);


-- +migrate Down
DROP TABLE recordings;
