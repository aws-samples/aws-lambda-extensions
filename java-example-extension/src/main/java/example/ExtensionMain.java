package example;

import com.google.gson.Gson;
import com.google.gson.JsonElement;
import com.google.gson.JsonObject;

/**
 * Entry point for external extension
 */
public class ExtensionMain {
    public static void main(String[] args) {
        // Register the extension for "INVOKE" and "SHUTDOWN" events
        final String extension = RegisterExtension.registerExtension();
        System.out.println("Extension registration complete, extensionID: " + extension);

        while (true) {
            try {
                String response = RegisterExtension.getNext(extension);
                if (response != null && !response.isEmpty()) {
                    JsonObject eventJsonObject = new Gson().fromJson(response, JsonObject.class);
                    JsonElement eventTypeElement = eventJsonObject.get("eventType");

                    // Depending upon event type perform corresponding actions
                    if (eventTypeElement != null) {
                        final String eventType = eventTypeElement.getAsString();
                        switch (eventTypeElement.getAsString()) {
                            case "INVOKE":
                                handleInvoke(response);
                                break;
                            case "SHUTDOWN":
                                handleShutDown();
                                break;
                            default:
                                System.err.println("Invalid event type received " + eventType);
                        }
                    }
                }
            } catch (Exception e) {
                System.out.println("Error while processing extension -" + e.getMessage());
                e.printStackTrace();
            }
        }
    }

    /**
     * Shutdown extension if we receive a shutdown event from lambda container
     */
    private static void handleShutDown() {
        System.out.println("Shutting down the extension");
        System.exit(0);
    }

    /**
     * Process payload
     *
     * @param payload event payload
     */
    public static void handleInvoke(String payload) {
        System.out.println("Handling invoke from extension " + payload);
    }
}