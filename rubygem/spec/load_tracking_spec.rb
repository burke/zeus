require 'zeus/load_tracking'

describe "Zeus::LoadTracking" do
  let(:test_filename) { __FILE__ }
  let(:test_dirname)  { File.dirname(test_filename) }

  class MyError < StandardError; end

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
        new_files, = Zeus::LoadTracking.features_loaded_by do
          $untracked_features << "an_untracked_feature.rb"
        end

        expect(new_files).to eq(["an_untracked_feature.rb"])
      end
    end
  end

  describe '.features_loaded_by' do
    def expect_to_load(expect_features, expect_err=NilClass)
      new_files, err = Zeus::LoadTracking.features_loaded_by do
        yield
      end

      expect(new_files.sort).to eq(expect_features.sort)
      expect(err).to be_instance_of(expect_err)
    end

    def expand_asset_path(path)
      File.join(__dir__, 'assets', path)
    end

    context 'loading valid code' do
      it 'tracks successful require_relative' do
        expect_to_load([expand_asset_path('require_relative.rb')]) do
          require_relative 'assets/require_relative'
        end
      end

      it 'tracks successful require' do
        expect_to_load([expand_asset_path('require.rb')]) do
          require expand_asset_path('require')
        end
      end

      it 'tracks loads' do
        expect_to_load([expand_asset_path('load.rb')]) do
          load expand_asset_path('load.rb')
        end
      end
    end

    context 'loading invalid code' do
      it 'tracks requires that raise a SyntaxError' do
        expect_to_load([expand_asset_path('invalid_syntax.rb')], SyntaxError) do
          require expand_asset_path('invalid_syntax')
        end
      end

      it 'tracks requires that raise a RuntimeError' do
        expect_to_load([expand_asset_path('runtime_error.rb')], RuntimeError) do
          require expand_asset_path('runtime_error')
        end
      end

      it 'tracks requires that exit' do
        expect_to_load([expand_asset_path('exit.rb')], SystemExit) do
          require expand_asset_path('exit')
        end
      end

      it 'tracks requires that throw in a method call' do
        expect_to_load([expand_asset_path('raise.rb')], MyError) do
          require expand_asset_path('raise')
          raise_it(MyError)
        end
      end
    end
  end
end
