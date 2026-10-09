/** Thrown when the translator meets a construct it cannot translate yet. */
final class Unsupported extends RuntimeException {
    Unsupported(String reason) {
        super(reason);
    }
}
