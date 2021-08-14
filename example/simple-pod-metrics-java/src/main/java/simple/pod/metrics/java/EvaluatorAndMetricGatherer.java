/*
Copyright 2021 The Custom Pod Autoscaler Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package simple.pod.metrics.java;

import java.io.BufferedReader;
import java.io.InputStreamReader;
import java.net.HttpURLConnection;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.net.URI;

import org.apache.commons.cli.Options;
import org.apache.commons.cli.CommandLineParser;
import org.apache.commons.cli.CommandLine;
import org.apache.commons.cli.DefaultParser;
import org.apache.commons.cli.Option;
import com.google.gson.JsonObject;
import com.google.gson.JsonParser;
import com.google.gson.JsonArray;

public class EvaluatorAndMetricGatherer {
    public static void main(String[] args) throws Exception {
        // Command line options for parsing execution mode
        Options options = new Options();

        Option input = new Option("m", "mode", true, "execution mode");
        input.setRequired(true);
        options.addOption(input);

        CommandLineParser parser = new DefaultParser();
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
        JsonObject jsonInput = JsonParser.parseString(stdinBuilder.toString()).getAsJsonObject();

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
            JsonObject value = JsonParser.parseString(metric.get("value").getAsString()).getAsJsonObject();
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
