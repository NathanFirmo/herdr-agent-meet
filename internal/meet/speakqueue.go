package meet

import "log/slog"

const speakQueueSize = 32

// speakJob is either paragraphs to summarize or a fixed line to say.
type speakJob struct {
	paragraphs []string
	text       string
}

// enqueueSpeech hands a job to the agent's own worker, so each agent speaks
// in order while different agents are generated in parallel.
func (r *Room) enqueueSpeech(agentID string, job speakJob) {
	r.mu.Lock()
	defer r.mu.Unlock()

	jobs, ok := r.speakQueues[agentID]
	if !ok {
		jobs = make(chan speakJob, speakQueueSize)
		r.speakQueues[agentID] = jobs
		go r.speakWorker(agentID, jobs)
	}
	select {
	case jobs <- job:
	default:
		slog.Warn("speech queue full, dropping", "agent", agentID)
	}
}

// speakWorker merges paragraphs that piled up while the previous speech was
// being generated into a single summary, so a busy agent does not fall behind.
func (r *Room) speakWorker(agentID string, jobs <-chan speakJob) {
	for job := range jobs {
		pending := []speakJob{job}
	drain:
		for {
			select {
			case next, ok := <-jobs:
				if !ok {
					break drain
				}
				pending = append(pending, next)
			default:
				break drain
			}
		}

		var paragraphs []string
		for _, job := range pending {
			if job.text != "" {
				if len(paragraphs) > 0 {
					r.speak(agentID, paragraphs, "")
					paragraphs = nil
				}
				r.speak(agentID, nil, job.text)
				continue
			}
			paragraphs = append(paragraphs, job.paragraphs...)
		}
		if len(paragraphs) > 0 {
			r.speak(agentID, paragraphs, "")
		}
	}
}
