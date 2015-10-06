require 'spec_helper'

describe "Zeus::LoadTracking" do
  let(:test_filename) { __FILE__ }
  let(:test_dirname)  { File.dirname(test_filename) }

  describe '.add_feature' do
    context 'already in load path' do
      before do
        # add the dir path of the tempfile to LOAD_PATH
        $LOAD_PATH  << test_dirname
      end

      after { $LOAD_PATH.delete test_dirname }


      it 'adds full filepath to $untracked_features' do
        Zeus::LoadTracking.add_feature(test_filename)

        expect($untracked_features).to include(__dir__ + "/load_tracking_spec.rb")
      end
    end

    context 'not in load path' do
      it 'adds full filepath to $untracked_features' do
        Zeus::LoadTracking.add_feature(test_filename)

        expect($untracked_features).to include(__dir__ + "/load_tracking_spec.rb")
      end
    end

    context '.features_loaded_by' do
      it 'returns list of new files loaded when block executes' do
        new_files = Zeus::LoadTracking.features_loaded_by do
          $untracked_features << "an_untracked_feature.rb"
        end

        expect(new_files).to eq(["an_untracked_feature.rb"])
      end
    end
  end
end
