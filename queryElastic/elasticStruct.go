package main

type ResponseStruct struct {
	Took     int          `json:"took"`
	TimedOut bool         `json:"timed_out"`
	Shards   ShardsStruct `json:"_shards"`
	Hits     HitsStruct   `json:"hits"`
}

type ShardsStruct struct {
	Total      int `json:"total"`
	Successful int `json:"successful"`
	Skipped    int `json:"skipped"`
	Failed     int `json:"failed"`
}

type HitsStruct struct {
	Total    TotalInfoStruct `json:"total"`
	MaxScore float64         `json:"max_score"`
	Hits     []HitStruct     `json:"hits"`
}

type TotalInfoStruct struct {
	Value    int    `json:"value"`
	Relation string `json:"relation"`
}

type HitStruct struct {
	Index  string       `json:"_index"`
	ID     string       `json:"_id"`
	Score  float64      `json:"_score"`
	Source SourceStruct `json:"_source"`
}

type SourceStruct struct {
	Metadata     MetadataStruct     `json:"metadata"`
	Process      ProcessStruct      `json:"process"`
	Log          LogStruct          `json:"log"`
	RealMessage  string             `json:"real_message"`
	ElasticAgent ElasticAgentStruct `json:"elastic_agent"`
	Syslog       SyslogStruct       `json:"syslog"`
	Message      string             `json:"message"`
	Tags         []string           `json:"tags"`
	Ingest       IngestStruct       `json:"ingest"`
	Input        InputStruct        `json:"input"`
	Timestamp    string             `json:"@timestamp"`
	ECS          ECSStruct          `json:"ecs"`
	DataStream   DataStreamStruct   `json:"data_stream"`
	Host         HostStruct         `json:"host"`
	Version      string             `json:"@version"`
	Event        EventStruct        `json:"event"`
}

type MetadataStruct struct {
	Pipeline  string      `json:"pipeline"`
	Input     InputStruct `json:"input"`
	RawIndex  string      `json:"raw_index"`
	StreamID  string      `json:"stream_id"`
	Beat      string      `json:"beat"`
	Truncated bool        `json:"truncated"`
	Type      string      `json:"type"`
	Version   string      `json:"version"`
	InputID   string      `json:"input_id"`
}

type ProcessStruct struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

type LogStruct struct {
	Source SourceAddressStruct `json:"source"`
}

type SourceAddressStruct struct {
	Address string `json:"address"`
}

type ElasticAgentStruct struct {
	ID       string `json:"id"`
	Version  string `json:"version"`
	Snapshot bool   `json:"snapshot"`
}

type SyslogStruct struct {
	Severity      int    `json:"severity"`
	Host          string `json:"host"`
	PID           int    `json:"pid"`
	Program       string `json:"program"`
	Priority      int    `json:"priority"`
	Facility      int    `json:"facility"`
	SeverityLabel string `json:"severity_label"`
	Timestamp     string `json:"timestamp"`
	FacilityLabel string `json:"facility_label"`
}

type IngestStruct struct {
	Timestamp string `json:"timestamp"`
}

type InputStruct struct {
	Type string `json:"type"`
}

type ECSStruct struct {
	Version string `json:"version"`
}

type DataStreamStruct struct {
	Namespace string `json:"namespace"`
	Type      string `json:"type"`
	Dataset   string `json:"dataset"`
}

type HostStruct struct {
	Hostname      string   `json:"hostname"`
	OS            OSStruct `json:"os"`
	IP            []string `json:"ip"`
	Containerized bool     `json:"containerized"`
	Name          string   `json:"name"`
	ID            string   `json:"id"`
	MAC           []string `json:"mac"`
	Architecture  string   `json:"architecture"`
}

type OSStruct struct {
	Kernel   string `json:"kernel"`
	Name     string `json:"name"`
	Family   string `json:"family"`
	Type     string `json:"type"`
	Version  string `json:"version"`
	Platform string `json:"platform"`
}

type EventStruct struct {
	Module  string `json:"module"`
	Dataset string `json:"dataset"`
}
