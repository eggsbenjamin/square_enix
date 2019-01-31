CREATE TABLE IF NOT EXISTS Process (
  id          INT PRIMARY KEY AUTO_INCREMENT,
  status      VARCHAR(50) NOT NULL,
  created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS Element (
  id          INT PRIMARY KEY AUTO_INCREMENT,
  data        VARCHAR(50),
  created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS ProcessElement (
  process_id INT,
  element_id INT,

  FOREIGN KEY(process_id) REFERENCES Process(id),
  FOREIGN KEY(element_id) REFERENCES Element(id)
);
