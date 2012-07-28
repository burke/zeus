module Zeus
  module Detector
    extend self

    DEFAULT = {
      type: :unknown
    }

    DETECTORS = [
      {
        file: 'script/rails',
        type: :rails
      },
      {
        file: 'config.ru',
        type: :rack
      }
    ]

    def from_detectors
      DETECTORS.detect do |detector|
        file = detector.fetch(:file)
        File.exists?(file)
      end
    end

    def detect
      detection = from_detectors || DEFAULT

      detection.fetch(:type)
    end
  end
end
