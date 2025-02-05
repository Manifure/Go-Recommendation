package kafka

type MockProducer struct {
	Messages []string
	Err      error
}

func (m *MockProducer) Produce(message string) error {
	if m.Err != nil {
		return m.Err
	}

	m.Messages = append(m.Messages, message)

	return nil
}
