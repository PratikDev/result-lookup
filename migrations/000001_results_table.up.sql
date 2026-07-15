DO $$ BEGIN
    CREATE TYPE board_name AS ENUM (
      'Dhaka',
      'Chattogram',
      'Rajshahi',
      'Cumilla',
      'Jessore',
      'Barishal',
      'Sylhet',
      'Dinajpur',
      'Mymensingh',
      'Madrasah'
    );
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

CREATE TABLE results (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    roll INT NOT NULL,
    reg INT NOT NULL,
    student_name VARCHAR(200) NOT NULL,
    institution_name VARCHAR(300) NOT NULL,
    board_name board_name NOT NULL,
    exam_year INT NOT NULL,
    gpa NUMERIC(3, 2) NOT NULL CHECK (gpa >= 0.00 AND gpa <= 5.00),
    is_passed BOOLEAN NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT unique_result UNIQUE (roll, reg, exam_year)
);