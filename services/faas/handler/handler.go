package handler

import (
    "context"
    "encoding/json"
    "fmt"
    "github.com/nats-io/nats.go"
    "github.com/go-redis/redis/v8"
)

var ctx = context.Background()
var redisClient *redis.Client
var natsConn *nats.Conn

func InitConnections() {

    //redisClient := redis.NewClient(&redis.Options{Addr: "redis:6379"})
    redisClient = redis.NewClient(&redis.Options{
        Addr: "redis:6379",
    })
    var err error
    natsConn, err = nats.Connect(nats.DefaultURL)
    if err != nil {
        panic(err)
    }
}

func RegisterFunction(name string, code string) string {
    funcKey := fmt.Sprintf("func:%s:code", name)
    luaCode := fmt.Sprintf("%s", code)
    
    err := redisClient.Set(ctx, funcKey, luaCode, 0).Err()
    if err != nil {
        return fmt.Sprintf(`{"status": "error", "message": "Failed to register function: %s"}`, err.Error())
    }

    return `{"status": "success", "message": "Function registered successfully"}`
}

func CallFunction(name string, async bool, replyTo string, args ...string) string {
    funcKey := fmt.Sprintf("func:%s:code", name)
    code, err := redisClient.Get(ctx, funcKey).Result()

    if err == redis.Nil {
        return `{"status": "error", "message": "Function not found"}`
    } else if err != nil {
        return fmt.Sprintf(`{"status": "error", "message": "Error retrieving function from Redis: %s"}`, err.Error())
    }

    redisArgs := make([]interface{}, len(args))
    for i, v := range args {
        redisArgs[i] = v
    }

    historyKey := fmt.Sprintf("func:%s:history", name)

    if async {
        if replyTo == "" {
            replyTo = fmt.Sprintf("response.%s", name)
        }
        go func() {
            result, err := redisClient.Eval(ctx, code, []string{}, redisArgs...).Result()
            if err != nil {
                errorMessage := fmt.Sprintf(`{"status": "error", "message": "Error executing function: %s"}`, err.Error())
                natsConn.Publish(replyTo, []byte(errorMessage))
            } else {
                redisClient.LPush(ctx, historyKey, fmt.Sprintf("%v", result))
                successMessage := fmt.Sprintf(`{"status": "success", "result": "%v"}`, result)
                natsConn.Publish(replyTo, []byte(successMessage))
            }
        }()
        return `{"status": "pending", "message": "Function execution in progress (async)"}`
    } else {
        result, err := redisClient.Eval(ctx, code, []string{}, redisArgs...).Result()
        if err != nil {
            return fmt.Sprintf(`{"status": "error", "message": "Error executing function: %s"}`, err.Error())
        }
        redisClient.LPush(ctx, historyKey, fmt.Sprintf("%v", result))
        return fmt.Sprintf(`{"status": "success", "result": "%v"}`, result)
    }
}

func DeregisterFunction(name string) string {
    codeKey := fmt.Sprintf("func:%s:code", name)
    historyKey := fmt.Sprintf("func:%s:history", name)

    deleted, err := redisClient.Del(ctx, codeKey).Result()
    redisClient.Del(ctx, historyKey).Result()

    if err != nil {
        return fmt.Sprintf(`{"status": "error", "message": "Failed to delete function: %s"}`, err.Error())
    }
    if deleted == 0 {
        return `{"status": "error", "message": "Function not found"}`
    }

    return `{"status": "success", "message": "Function deleted successfully"}`
}

func ListFunctions() string {
    keys, _ := redisClient.Keys(ctx, "func:*:code").Result()
    for i := range keys {
        keys[i] = keys[i][5 : len(keys[i])-5]
    }

    jsonData, _ := json.Marshal(map[string]interface{}{
        "status": "success",
        "functions": keys,
    })
    return string(jsonData)
}

func GetFunctionHistory(name string) string {
    historyKey := fmt.Sprintf("func:%s:history", name)
    history, err := redisClient.LRange(ctx, historyKey, 0, -1).Result()
    if err != nil {
        return `{"status": "error", "message": "Error retrieving function history"}`
    }

    jsonData, _ := json.Marshal(map[string]interface{}{
        "status": "success",
        "history": history,
    })
    return string(jsonData)
}

func SubscribeInvoke() {
    natsConn.QueueSubscribe("invoke.>", "function_workers", func(msg *nats.Msg) {
        functionName := msg.Subject[len("invoke."):]

        var payload struct {
            Args    []string `json:"args"`
            Async   bool     `json:"async"`
            ReplyTo string   `json:"reply_to"`
        }

        err := json.Unmarshal(msg.Data, &payload)
        if err != nil {
            fmt.Println("Error parsing JSON request:", err)
            return
        }

        result := CallFunction(functionName, payload.Async, payload.ReplyTo, payload.Args...)
        if !payload.Async {
            fmt.Println("Function Response:", result)
        }
    })
    select {}
}
