package simple.pod.metrics.java;

import java.io.*;
import java.net.HttpURLConnection;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.net.URI;

import org.apache.commons.cli.*;
import com.google.gson.*;

public class EvaluatorAndMetricGatherer {
    public static void main(String[] args) throws Exception {
        // Command line options for parsing execution mode
        Options options = new Options();

        Option input = new Option("m", "mode", true, "execution mode");
        input.setRequired(true);
        options.addOption(input);

        CommandLineParser parser = new DefaultParser();
        HelpFormatter formatter = new HelpFormatter();
        CommandLine cmd = parser.parse(options, args);

        // Get execution mode
        String mode = cmd.getOptionValue("mode");

        // Read stdin into a string
        BufferedReader in = new BufferedReader(new InputStreamReader(System.in));
        StringBuilder stdinBuilder = new StringBuilder();
        String s;
        while ((s = in.readLine()) != null && s.length() != 0) {
            stdinBuilder.append(s);
        }

        // Convert stdin into a JSON object
        JsonObject jsonInput = new JsonParser().parse(stdinBuilder.toString()).getAsJsonObject();

        // Handle execution differently based on the mode provided, either metric gathering or evaluation
        switch(mode) {
            case "evaluate":
                evaluate(jsonInput);
                break;
            case "metric":
                metric(jsonInput);
                break;
            default:
                System.err.printf("Unknown execution mode '%s'", mode);
                System.exit(1);
        }
    }

    private static void evaluate(JsonObject input) {
        int totalAvailable = 0;
        JsonArray metrics = input.get("metrics").getAsJsonArray();
        for (int i = 0; i < metrics.size(); i++) {
            JsonObject metric = metrics.get(i).getAsJsonObject();
            JsonObject value = new JsonParser().parse(metric.get("value").getAsString()).getAsJsonObject();
            int available = value.get("available").getAsInt();
            totalAvailable += available;
        }

        int targetReplicaCount = input.get("resource").getAsJsonObject()
                                      .get("spec").getAsJsonObject()
                                      .get("replicas").getAsInt();

        if (totalAvailable > 5) {
            targetReplicaCount -= 1;
        }

        if (totalAvailable <= 0) {
            targetReplicaCount += 1;
        }

        System.out.printf("{ \"targetReplicas\" : %d }", targetReplicaCount);
    }

    private static void metric(JsonObject input) throws Exception {
        String ip = input.get("resource").getAsJsonObject()
                         .get("status").getAsJsonObject()
                         .get("podIP").getAsString();

        HttpClient client = HttpClient.newHttpClient();
        HttpRequest request = HttpRequest.newBuilder()
                                         .uri(URI.create(String.format("http://%s:5000/metric", ip)))
                                         .build();
        HttpResponse<String> response = client.send(request, HttpResponse.BodyHandlers.ofString());
        if (response.statusCode() != HttpURLConnection.HTTP_OK) {
            throw new Exception(
                String.format("Error occurred retrieving metrics, got response status: %d", response.statusCode()));
        }
        System.out.print(response.body());
    }
}
