require 'zeus/load_tracking'

describe "Zeus::LoadTracking" do
  let(:test_filename) { __FILE__ }
  let(:test_dirname)  { File.dirname(test_filename) }

  class MyError < StandardError; end

  def expect_to_load(expect_features, expect_err=NilClass)
    buf = StringIO.new
    Zeus::LoadTracking.set_feature_pipe(buf)

    begin
      Zeus::LoadTracking.track_features_loaded_by do
        yield
      end
    rescue ScriptError => err
    rescue => err
    end

    expect(err).to be_instance_of(expect_err)
    expect(buf.string.strip.split("\n").sort).to eq(expect_features.sort)
  end

  def expand_asset_path(path)
    File.join(__dir__, 'assets', path)
  end

  describe '.add_feature' do
    it 'tracks full filepath' do
      relative_path = Pathname.new(test_filename).relative_path_from(Pathname.new(Dir.pwd)).to_s

      expect_to_load([test_filename]) do
        Zeus::LoadTracking.add_feature(relative_path)
      end
    end

    it 'tracks loads' do
      target = expand_asset_path('load.rb')

      # The first `load` in Travis also pulls enc/trans/single_byte.so from the
      # Ruby VM for some reason. Loading twice is harmless since this file is empty.
      load(target)

      expect_to_load([target]) do
        load(target)
      end
    end

    it 'does not error outside a tracking block without Zeus configured' do
      Zeus::LoadTracking.add_feature(test_filename)
    end
  end

  describe '.track_features_loaded_by' do
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
    end

    context 'loading invalid code' do
      it 'tracks requires that raise a SyntaxError' do
        expect_to_load([test_filename, expand_asset_path('invalid_syntax.rb')], SyntaxError) do
          require expand_asset_path('invalid_syntax')
        end
      end

      it 'tracks requires that raise a RuntimeError' do
        expect_to_load([test_filename, expand_asset_path('runtime_error.rb')], RuntimeError) do
          require expand_asset_path('runtime_error')
        end
      end

      it 'tracks requires that throw in a method call' do
        expect_to_load([test_filename, expand_asset_path('raise.rb')], MyError) do
          require expand_asset_path('raise')
          raise_it(MyError)
        end
      end
    end
  end
end
